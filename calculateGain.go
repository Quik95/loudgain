package loudgain

import (
	"math"
)

// CalculateTrackGain applies gain based on measured loudness of the audio file relative to the reference level.
func CalculateTrackGain(in float64) float64 {
	const referenceLevel = -18

	return referenceLevel - in
}

// DecibelsToLinear converts loudness in the dBTP unit to the linear 0..1 scale.
func DecibelsToLinear(in float64) float64 {
	return math.Pow(10, in/20)
}

// LinearToDecibels converts loudness in the linear scale to the dBTP unit.
func LinearToDecibels(in float64) float64 {
	return math.Log10(in) * 20
}

// PreventClippint checks if after applying gain the clipping will occur, and lowers a track's peak if necessary.
func PreventClippint(ll LoudnessLevel) float64 {
	const pregain = 0

	trackPeakLimit := DecibelsToLinear(-1.0)
	trackGain := CalculateTrackGain(ll.IntegratedLoudness) + pregain
	trackPeakAfterGain := DecibelsToLinear(trackGain) * DecibelsToLinear(ll.TruePeakdB)

	if trackPeakAfterGain > trackPeakLimit {
		return trackGain - LinearToDecibels(trackPeakAfterGain/trackPeakLimit)
	}

	return trackGain
}
