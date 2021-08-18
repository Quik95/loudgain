package loudgain

import (
	"math"
)

// CalculateTrackGain applies gain based on measured loudness of the audio file relative to the reference level.
func CalculateTrackGain(in LoudnessUnit) LoudnessUnit {
	const referenceLevel = -18

	return referenceLevel - in
}

// ToLinear converts loudness in the dBTP unit to the linear 0..1 scale.
func (d Decibel) ToLinear() LinearLoudness {
	return LinearLoudness(math.Pow(10, float64(d)/20))
}

// ToLoudnessUnit converts Decibel to the LoudnessUnit.
func (d Decibel) ToLoudnessUnit() LoudnessUnit {
	return LoudnessUnit(d)
}

// ToDecibels converts loudness in the linear scale to the dBTP unit.
func (l LinearLoudness) ToDecibels() Decibel {
	return Decibel(math.Log10(float64(l)) * 20)
}

// ToLoudnessUnit converts LinearLoudness to the LoudnessUnit.
func (l LinearLoudness) ToLoudnessUnit() LoudnessUnit {
	return l.ToDecibels().ToLoudnessUnit()
}

// ToDecibels converts LoudnessUnit to the Decibel.
func (l LoudnessUnit) ToDecibels() Decibel {
	return Decibel(l)
}

// ToLinear converts LoudnessUnit to the LinearLoudness.
func (l LoudnessUnit) ToLinear() LinearLoudness {
	return l.ToDecibels().ToLinear()
}

// PreventClippint checks if after applying gain the clipping will occur, and lowers a track's peak if necessary.
func PreventClippint(ll LoudnessLevel) LoudnessUnit {
	const pregain = 0

	trackPeakLimit := LoudnessUnit(-1.0).ToLinear()
	trackGain := CalculateTrackGain(ll.IntegratedLoudness) + pregain
	trackPeakAfterGain := trackGain.ToLinear() * ll.TruePeakdB.ToLinear()

	if trackPeakAfterGain > trackPeakLimit {
		return trackGain - (trackPeakAfterGain / trackPeakLimit).ToLoudnessUnit()
	}

	return trackGain
}
