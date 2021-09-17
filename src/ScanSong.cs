using System;
using System.Diagnostics;
using System.Text;
using System.Threading.Tasks;
using Xabe.FFmpeg;

namespace loudgain
{
    public record ReplaygainValues
    {
        public Decibel Gain { get; init; }
        public LinearLoudness Peak { get; init; }
        public Decibel Range { get; init; }
        public LoudnessUnitFullScale Reference { get; init; }

        public ReplaygainValues(Decibel gain, LinearLoudness peak, Decibel range, LoudnessUnitFullScale reference)
        {
            this.Gain = gain;
            this.Peak = peak;
            this.Range = range;
            this.Reference = reference;
        }
    }

    public record TrackValues : ReplaygainValues
    {
        public TrackValues(Decibel gain, LinearLoudness peak, Decibel range, LoudnessUnitFullScale reference) : base(
            gain, peak, range, reference)
        {
        }
    }

    public record AlbumValues : ReplaygainValues
    {
        public AlbumValues(Decibel gain, LinearLoudness peak, Decibel range, LoudnessUnitFullScale reference) : base(
            gain, peak, range, reference)
        {
        }
    }

    public class ScanResult
    {
        public string FilePath { get; set; }
        public TrackValues? Track { get; set; }
        public AlbumValues? Album { get; set; }

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

        public static async Task<TrackValues?> TrackScan(string song)
        {
            var args =
                $"-i \"{song}\" -hide_banner -nostats -filter_complex ebur128=peak='true':framelog='verbose' -f null -";

            var conversion = FFmpeg.Conversions.New();

            var parser = new FFmpegOutputParser();
            
            conversion.OnDataReceived += parser.ConversionOnOnDataReceived;
            
            var conversionResult = await conversion.Start(args);
            
            Console.WriteLine(parser.GetFFmpegOutput());

            return null;
        }
    }

    public class FFmpegOutputParser
    {
        private StringBuilder _ffmpegOutput;

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
    }
}