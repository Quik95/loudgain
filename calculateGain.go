package loudgain

import (
	"fmt"
	"math"
)

// CalculateTrackGain applies gain based on measured loudness of the audio file relative to the reference level.
func CalculateTrackGain(loudness LoudnessUnit) LoudnessUnit {
	return ReferenceLoudness - loudness + Pregain
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

// PreventClipping checks if after applying gain the clipping will occur, and lowers a track's peak if necessary.
func PreventClipping(trackPeak Decibel, trackGain LoudnessUnit) LoudnessUnit {
	trackPeakAfterGain := trackGain.ToLinear() * trackPeak.ToLinear()

	if trackPeakAfterGain > TrackPeakLimit.ToLinear() {
		return trackGain - (trackPeakAfterGain / TrackPeakLimit.ToLinear()).ToLoudnessUnit()
	}

	return trackGain
}

// ScanResult contains the results of scanning an audio file and applying gain to it.
type ScanResult struct {
	FilePath                    string
	TrackGain, TrackRange       Decibel
	ReferenceLoudness, Loudness LoudnessUnit
	TrackPeak                   LinearLoudness
}

func (s ScanResult) String() string {
	return fmt.Sprintf(
		"Filepath: %s\n"+
			"Loudness: %8.2f LUFS\n"+
			"Range: %12s\n"+
			"Peak: %14f (%f dBTP)\n"+
			"Gain: %14s\n",
		s.FilePath, s.Loudness, s.TrackRange, s.TrackPeak, s.TrackPeak.ToDecibels(), s.TrackGain)
}
