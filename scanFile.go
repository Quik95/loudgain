package loudgain

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"
)

func CheckExtension(filepath string) error {
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

	if _, ok := allowed[extension]; !ok {
		return fmt.Errorf("unsupported file format in song: %s", filepath)
	}

	return nil
}

func ScanFile(filepath string) ScanResult {
	if err := CheckExtension(filepath); err != nil {
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

	trackGain := CalculateTrackGain(ll.IntegratedLoudness)
	if NoClip {
		trackGain = PreventClipping(ll.TruePeakdB, trackGain)
	}

	res := ScanResult{
		FilePath:          filepath,
		TrackGain:         trackGain.ToDecibels(),
		TrackRange:        ll.LoudnessRange.ToDecibels(),
		ReferenceLoudness: ReferenceLoudness,
		TrackPeak:         ll.TruePeakdB.ToLinear(),
		Loudness:          ll.IntegratedLoudness,
	}

	return res
}

func GetScannedSongs(songs []string) <-chan ScanResult {
	var wg sync.WaitGroup

	wg.Add(len(songs))

	guard := make(chan struct{}, WorkersLimit)
	reschan := make(chan ScanResult, len(songs))
	pg := GetProgressBar(len(songs))
	pg.Describe("Scanning tracks")

	doScan := func(song string) {
		reschan <- ScanFile(song)

		wg.Done()
		pg.Add(1)
		<-guard
	}

	for _, song := range songs {
		guard <- struct{}{}

		go doScan(song)
	}

	wg.Wait()
	close(reschan)

	return reschan
}
