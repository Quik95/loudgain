using System;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using Xabe.FFmpeg;

namespace loudgain
{
    public record ReplaygainValues(Decibel Gain, LinearLoudness Peak, Decibel Range, LoudnessUnitFullScale Reference,
        LoudnessUnitFullScale Loudness);

    public class ScanResult
    {
        public string FilePath { get; set; }
        public ReplaygainValues? Track { get; set; }
        public ReplaygainValues? Album { get; set; }

        public ScanResult(string path)
        {
            this.FilePath = path;
            this.Track = null;
            this.Album = null;

            /* this.Track = new TrackValues( */
            /*     new Decibel(10), */
            /*     new LinearLoudness(10), */
            /*     new Decibel(10), */
            /*     new LoudnessUnitFullScale(20) */
            /* ); */


            /* this.Album = new AlbumValues( */
            /*     new Decibel(10), */
            /*     new LinearLoudness(10), */
            /*     new Decibel(10), */
            /*     new LoudnessUnitFullScale(20) */
            /* ); */
        }

        public override string ToString()
        {
            return $"File: {this.FilePath}\n{this.Track}\n{this.Album}";
        }

        public static async Task<ReplaygainValues?> TrackScan(string song)
        {
            var args =
                $"-i \"{song}\" -hide_banner -nostats -filter_complex ebur128=peak='true':framelog='verbose' -f null -";

            var conversion = FFmpeg.Conversions.New();
            var parser = new FFmpegOutputParser();

            conversion.OnDataReceived += parser.ConversionOnOnDataReceived;

            await conversion.Start(args);

            var values = parser.GetMatches();
            if (values is null)
            {
                Console.WriteLine(parser.GetFFmpegOutput());
                return null;
            }

            var gain = Gain.CalculateGain(values.IntegratedLoudness, values.Peak);
            var res = new ReplaygainValues(gain, values.Peak.ToLinear(), values.LoudnessRange.ToDecibel(),
                new LoudnessUnitFullScale(-18), values.IntegratedLoudness);

            return res;
        }
    }

    public class FFmpegOutputParser
    {
        private readonly StringBuilder _ffmpegOutput;

        private static readonly Regex IntegratedLoudnessRegexp =
            new Regex(@"I:\s*(?<value>-?\d+\.?\d{1})\sLUFS", RegexOptions.Compiled);

        private static readonly Regex LoudnessRangeRegexp =
            new Regex(@"LRA:\s*(?<value>-?\d+\.?\d{1})\sLU", RegexOptions.Compiled);

        private static readonly Regex TruePeakRegexp =
            new Regex(@"Peak:\s*(?<value>-?\d+\.?\d{1})\sdBFS", RegexOptions.Compiled);

        public FFmpegOutputParser()
        {
            this._ffmpegOutput = new StringBuilder();
        }

        public void ConversionOnOnDataReceived(object sender, DataReceivedEventArgs e)
        {
            this._ffmpegOutput.AppendLine(e.Data);
        }

        public string GetFFmpegOutput()
        {
            return this._ffmpegOutput.ToString();
        }

        public record ScanValues(LoudnessUnitFullScale IntegratedLoudness, Decibel Peak,
            LoudnessUnit LoudnessRange);

        public ScanValues? GetMatches()
        {
            var searchString = this.GetFFmpegOutput();

            var integratedLoudnessMatches = IntegratedLoudnessRegexp.Matches(searchString);
            if (integratedLoudnessMatches.Count == 0)
                return null;

            var integratedLoudnessString = integratedLoudnessMatches.Last().Groups["value"].Value;
            if (!LoudnessUnitFullScale.TryParse(integratedLoudnessString, out LoudnessUnitFullScale integratedLoudness))
                return null;

            var peakMatches = TruePeakRegexp.Matches(searchString);
            if (peakMatches.Count == 0)
                return null;

            var peakString = peakMatches.Last().Groups["value"].Value;
            if (!Decibel.TryParse(peakString, out Decibel peak))
                return null;

            var loudnessRangeMatches = LoudnessRangeRegexp.Matches(searchString);
            if (loudnessRangeMatches.Count == 0)
                return null;

            var loudnessRangeString = loudnessRangeMatches.Last().Groups["value"].Value;
            if (!LoudnessUnit.TryParse(loudnessRangeString, out LoudnessUnit loudnessRange))
                return null;

            return new ScanValues(integratedLoudness, peak, loudnessRange);
        }
    }
}