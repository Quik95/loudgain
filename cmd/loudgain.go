package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Quik95/loudgain"
)

var (
	flagPeakLimit, flagPregain float64
)

func init() {
	flag.Float64Var(&flagPeakLimit, "maxtpl", -1.0, "Maximal true peak level in dBTP")
	flag.Float64Var(&flagPregain, "pregain", 0.0, "Apply n dB/LU pre-gain value")

	flag.Parse()
}

func main() {
	var (
		referenceLoudness loudgain.LoudnessUnit = -18
		trackPeakLimit                          = loudgain.Decibel(flagPeakLimit)
		pregain                                 = loudgain.LoudnessUnit(flagPregain)
	)

	songs := flag.Args()
	if len(songs) == 0 {
		fmt.Println("No files to process. Exitting...")
		os.Exit(1)
	}

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
