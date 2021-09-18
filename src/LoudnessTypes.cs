using System;

namespace loudgain
{
    public class Decibel
    {
        public double Value { get; }

        public Decibel(double v)
        {
            this.Value = v;
        }

        public static bool TryParse(string s, out Decibel res)
        {
            if (!double.TryParse(s, out double temp))
            {
                res = new Decibel(0);
                return false;
            }

            res = new Decibel(temp);
            return true;
        }

        public override string ToString()
        {
            return $"{this.Value:F2} dB";
        }

        public LinearLoudness ToLinear()
        {
            return new LinearLoudness(Math.Pow(10, this.Value / 20));
        }
        
        public static Decibel operator -(Decibel self, Decibel other) =>
            new Decibel(self.Value - other.Value);
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

        public static bool TryParse(string s, out LoudnessUnit res)
        {
            if (!double.TryParse(s, out double temp))
            {
                res = new LoudnessUnit(0);
                return false;
            }

            res = new LoudnessUnit(temp);
            return true;
        }

        public LinearLoudness ToLinear()
        {
            return new LinearLoudness(Math.Pow(10, this.Value / 20));
        }

        public Decibel ToDecibel()
        {
            return new Decibel(this.Value);
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

        public static bool TryParse(string s, out LoudnessUnitFullScale res)
        {
            if (!double.TryParse(s, out double temp))
            {
                res = new LoudnessUnitFullScale(0);
                return false;
            }

            res = new LoudnessUnitFullScale(temp);
            return true;
        }

        public LinearLoudness ToLinear()
        {
            return new LinearLoudness(Math.Pow(10, this.Value / 20));
        }

        public static LoudnessUnitFullScale operator -(LoudnessUnitFullScale self, LoudnessUnitFullScale other) =>
            new LoudnessUnitFullScale(self.Value - other.Value);

        public static Decibel operator +(LoudnessUnitFullScale self, Decibel other) =>
            new Decibel(self.Value + other.Value);
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

        public Decibel ToDecibel()
        {
            return new Decibel(Math.Log10(this.Value) * 20);
        }
        
        public static LinearLoudness operator *(LinearLoudness self, LinearLoudness other) => new LinearLoudness(self.Value * other.Value);
        public static LinearLoudness operator /(LinearLoudness self, LinearLoudness other) => new LinearLoudness(self.Value / other.Value);
        public static bool operator >(LinearLoudness self, LinearLoudness other) => self.Value > other.Value;
        public static bool operator <(LinearLoudness self, LinearLoudness other) => self.Value < other.Value;
    }
}