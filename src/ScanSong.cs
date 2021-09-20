using System;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;

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
        }

        public override string ToString()
        {
            string trackString = "";
            string albumString = "";

            if (this.Track is null)
                return "";
            
            trackString = $"Track: {this.FilePath}\n" +
                          $"{"Loudness:",-10}{this.Track.Loudness}\n" +
                          $"{"Range:",-10}{this.Track.Range}\n" +
                          $"{"Peak:",-10}{this.Track.Peak} ({this.Track.Peak.ToDecibel()})\n" +
                          $"{"Gain:",-10}{this.Track.Gain}\n";

            if (this.Album is not null)
            {
                albumString = $"\nAlbum:\n" +
                              $"{"Loudness:",-10}{this.Album.Loudness}\n" +
                              $"{"Range:",-10}{this.Album.Range}\n" +
                              $"{"Peak:",-10}{this.Album.Peak} ({this.Album.Peak.ToDecibel()})\n" +
                              $"{"Gain:",-10}{this.Album.Gain}\n";
            }

            return trackString + albumString;
        }
        
        public static async Task<ReplaygainValues?> TrackScan(string song)
        {
            var args =
                $"-i \"{song}\" -hide_banner -nostats -filter_complex ebur128=peak='true':framelog='verbose' -f null -";

            try
            {
                using var process = new Process();
                process.StartInfo.UseShellExecute = false;
                process.StartInfo.FileName = "ffmpeg";
                process.StartInfo.Arguments = args;
                process.StartInfo.CreateNoWindow = true;
                process.StartInfo.RedirectStandardError = true;
                process.StartInfo.RedirectStandardOutput = true;
                process.Start();

                var ffmpegOutput = process.StandardError.ReadToEndAsync();
                await process.WaitForExitAsync();
                var parser = new FFmpegOutputParser();

                var values = parser.GetMatches(await ffmpegOutput);
                if (values is null)
                {
                    Program.MasterProgressBar.WriteErrorLine($"failed to scan song: {song}");
                    return null;
                }

                var gain = Gain.CalculateGain(values.IntegratedLoudness, values.Peak);
                var res = new ReplaygainValues(gain, values.Peak.ToLinear(), values.LoudnessRange.ToDecibel(),
                    new LoudnessUnitFullScale(-18), values.IntegratedLoudness);

                return res;
            }
            catch
            {
                    Program.MasterProgressBar.WriteErrorLine($"failed to scan song: {song}");
                return null;
            }
        }
    }

    public class FFmpegOutputParser
    {
        private static readonly Regex IntegratedLoudnessRegexp =
            new Regex(@"I:\s*(?<value>-?\d+\.?\d{1})\sLUFS", RegexOptions.Compiled);

        private static readonly Regex LoudnessRangeRegexp =
            new Regex(@"LRA:\s*(?<value>-?\d+\.?\d{1})\sLU", RegexOptions.Compiled);

        private static readonly Regex TruePeakRegexp =
            new Regex(@"Peak:\s*(?<value>-?\d+\.?\d{1})\sdBFS", RegexOptions.Compiled);

        public record ScanValues(LoudnessUnitFullScale IntegratedLoudness, Decibel Peak,
            LoudnessUnit LoudnessRange);

        public ScanValues? GetMatches(string searchString)
        {
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