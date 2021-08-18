package main

import (
	"log"
	"os"

	"github.com/Quik95/loudgain"
)

func main() {
	songs := os.Args[1:]
	filepath := songs[0]
	log.Println(songs)

	ffmpegPath, err := loudgain.GetFFmpegPath()
	if err != nil {
		log.Fatalln("ffmpeg not found in path")
	}

	log.Printf("ffmpeg is located at: %s", ffmpegPath)

	loudness, err := loudgain.RunLoudnessScan(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("%s not found\n", filepath)
		} else {
			log.Fatalf("an unknown error occurred while processing song %s\n", filepath)
		}
	}

	ll, err := loudgain.ParseLoudnessOutput(loudness, filepath)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Track Gain: %f", loudgain.PreventClippint(ll))
	log.Printf("Track Peak: %f (%f dBFS)", ll.TruePeakdB.ToLinear(), ll.TruePeakdB)
	log.Printf("Track Range: %f", ll.LoudnessRange)
}
