package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Quik95/loudgain"
)

func main() {
	var (
		referenceLoudness loudgain.LoudnessUnit = -18
		trackPeakLimit    loudgain.LoudnessUnit = -1
		pregain           loudgain.LoudnessUnit = 0
	)

	songs := os.Args[1:]
	filepath := songs[0]

	ffmpegPath, err := loudgain.GetFFmpegPath()
	if err != nil {
		log.Fatalln("an ffmpeg binary not found in the path")
	}

	log.Printf("the ffmpeg binary is located at: %s", ffmpegPath)

	loudness, err := loudgain.RunLoudnessScan(ffmpegPath, filepath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("%s not found\n", filepath)
		} else {
			log.Fatalf("an unknown error has occurred while processing song %s\n", filepath)
		}
	}

	ll, err := loudgain.ParseLoudnessOutput(loudness, filepath)
	if err != nil {
		log.Fatalln(err)
	}

	trackGain := loudgain.CalculateTrackGain(ll.IntegratedLoudness, referenceLoudness, pregain)
	trackGain = loudgain.PreventClipping(ll.TruePeakdB, trackGain, trackPeakLimit)

	res := loudgain.ScanResult{
		FilePath:          filepath,
		TrackGain:         trackGain.ToDecibels(),
		TrackRange:        ll.LoudnessRange.ToDecibels(),
		ReferenceLoudness: loudgain.LoudnessUnit(referenceLoudness),
		TrackPeak:         ll.TruePeakdB.ToLinear(),
		Loudness:          ll.IntegratedLoudness,
	}

	fmt.Println(res)
	if err := loudgain.WriteMetadata(ffmpegPath, res); err != nil {
		log.Println(err)
	}
}
