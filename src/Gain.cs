namespace loudgain
{
    public class Gain
    {
        public static Decibel CalculateGain(LoudnessUnitFullScale loudness, Decibel peak)
        {
            const bool noclip = false;
            var trackPeakLimit = new Decibel(-1);
            
            var gain = _calculateGain(loudness);
            if (!noclip)
                return gain;

            var trackPeakAfterGain = gain.ToLinear() * peak.ToLinear();

            if (trackPeakAfterGain > trackPeakLimit.ToLinear())
            {
                return gain - (trackPeakAfterGain / trackPeakLimit.ToLinear()).ToDecibel();
            }
            
            return gain;
        }

        private static Decibel _calculateGain(LoudnessUnitFullScale loudness)
        {
            var referenceLoudness = new LoudnessUnitFullScale(-18);
            var pregain  = new Decibel(0);
            
            return referenceLoudness - loudness + pregain;
        }
    }
}