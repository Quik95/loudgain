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
    }
}
