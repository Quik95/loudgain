namespace loudgain
{
    public class Decibel
    {
        public double Value { get; }

        public Decibel(double v)
        {
            this.Value = v;
        }

        public override string ToString()
        {
            return $"{this.Value:F2} dB";
        }
    }

    public class LoudnessUnit
    {
        public double Value { get; }

        public LoudnessUnit(double v)
        {
            this.Value = v;
        }

        public override string ToString()
        {
            return $"{this.Value:F2} LU";
        }
    }

    public class LoudnessUnitFullScale
    {
        public double Value { get; }

        public LoudnessUnitFullScale(double v)
        {
            this.Value = v;
        }

        public override string ToString()
        {
            return $"{this.Value:F2} LUFS";
        }
    }

    public class LinearLoudness
    {
        public double Value { get; }

        public LinearLoudness(double v)
        {
            this.Value = v;
        }

        public override string ToString()
        {
            return $"{this.Value:F7}";
        }
    }
}