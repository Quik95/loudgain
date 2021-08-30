package loudgain

import (
	"fmt"
	"log"
	"os"
	"path"
)

func checkExtension(filepath string) error {
	allowed := map[string]bool{
		".aiff": true,
		".aif":  true,
		".alfc": true,
		".ape":  true,
		".apl":  true,
		".bwf":  true,
		".flac": true,
		".mp3":  true,
		".mp4":  true,
		".m4a":  true,
		".m4b":  true,
		".m4p":  true,
		".m4r":  true,
		".mpc":  true,
		".ogg":  true,
		".tta":  true,
		".wma":  true,
		".wv":   true,
	}

	extension := path.Ext(filepath)

	if _, ok := allowed[extension]; ok != true {
		return fmt.Errorf("unsupported file format: %s", extension)
	}

	return nil
}

func ScanFile(filepath string) ScanResult {
	if err := checkExtension(filepath); err != nil {
		log.Println(err)

		return ScanResult{}
	}

	loudness, err := RunLoudnessScan(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("%s not found\n", filepath)
		} else {
			log.Println(err)
		}

		return ScanResult{}
	}

	ll, err := ParseLoudnessOutput(loudness, filepath)
	if err != nil {
		log.Println(err)

		return ScanResult{}
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

			return ScanResult{}
		}
	}

	return res
}
