package loudgain

import (
	"log"
	"os"
)

func ScanFile(filepath string) ScanResult {
	loudness, err := RunLoudnessScan(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("%s not found\n", filepath)
		} else {
			log.Println(err)
		}
	}

	ll, err := ParseLoudnessOutput(loudness, filepath)
	if err != nil {
		log.Println(err)
	}

	trackGain := CalculateTrackGain(ll.IntegratedLoudness, ReferenceLoudness, Pregain)
	if NoClip {
		trackGain = PreventClipping(ll.TruePeakdB, trackGain, TrackPeakLimit)
	}

	res := ScanResult{
		FilePath:          filepath,
		TrackGain:         trackGain.ToDecibels(),
		TrackRange:        ll.LoudnessRange.ToDecibels(),
		ReferenceLoudness: ReferenceLoudness,
		TrackPeak:         ll.TruePeakdB.ToLinear(),
		Loudness:          ll.IntegratedLoudness,
	}

	if TagMode != SkipWritingTags {
		if err := WriteMetadata(FFmpegPath, res, TagMode); err != nil {
			log.Println(err)
		}
	}

	return res
}
