namespace loudgain
{
    public class Gain
    {
        public static Decibel CalculateGain(LoudnessUnitFullScale loudness, Decibel peak)
        {
            var gain = _calculateGain(loudness);
            if (!Program.Config.Noclip)
                return gain;

            var trackPeakAfterGain = gain.ToLinear() * peak.ToLinear();

            if (trackPeakAfterGain > Program.Config.MaxTruePeakLevel.ToLinear())
            {
                return gain - (trackPeakAfterGain / Program.Config.MaxTruePeakLevel.ToLinear()).ToDecibel();
            }
            
            return gain;
        }

        private static Decibel _calculateGain(LoudnessUnitFullScale loudness)
        {
            var referenceLoudness = new LoudnessUnitFullScale(-18);
            var pregain = Program.Config.Pregain;
            
            return referenceLoudness - loudness + pregain;
        }
    }
}