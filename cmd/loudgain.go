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
	noClip                     bool
	tagMode                    string
)

func init() {
	flag.Float64Var(&flagPeakLimit, "maxtpl", -1.0, "Maximal true peak level in dBTP")
	flag.Float64Var(&flagPregain, "pregain", 0.0, "Apply n dB/LU pre-gain value")

	flag.BoolVar(&noClip, "noclip", false, "Lower track gain to avoid clipping.")
	flag.StringVar(&tagMode, "tagmode", "s",
		"--tagmode=d Delete ReplayGain tags from files. (uninmplemented)\n"+
			"--tagmode=i Write Replaygain 2.0 tags to files.\n"+
			"--tagmode=e like --tagmode=i, plus extra tags (reference, ranges).\n"+
			"--tagmode=l like --tagmode=e, but LU units instead of dB.\n"+
			"--tagmode=s Don't write Replaygain tags.")

	flag.Parse()
}

func checkExitCondition(tagMode loudgain.WriteMode) {
	if flag.NArg() == 0 {
		fmt.Println("No files to process. Exitting...")
		os.Exit(1)
	}

	if tagMode == loudgain.InvalidWriteMode {
		fmt.Println("Invalid write mode. Exitting...")
		os.Exit(1)
	}
}

func main() {
	var (
		referenceLoudness loudgain.LoudnessUnit = -18
		trackPeakLimit                          = loudgain.Decibel(flagPeakLimit)
		pregain                                 = loudgain.LoudnessUnit(flagPregain)
		tagMode                                 = loudgain.StringToWriteMode(tagMode)
	)

	checkExitCondition(tagMode)

	songs := flag.Args()
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
	if noClip {
		trackGain = loudgain.PreventClipping(ll.TruePeakdB, trackGain, trackPeakLimit)
	}

	res := loudgain.ScanResult{
		FilePath:          filepath,
		TrackGain:         trackGain.ToDecibels(),
		TrackRange:        ll.LoudnessRange.ToDecibels(),
		ReferenceLoudness: referenceLoudness,
		TrackPeak:         ll.TruePeakdB.ToLinear(),
		Loudness:          ll.IntegratedLoudness,
	}

	fmt.Println(res)

	if tagMode != loudgain.SkipWritingTags {
		if err := loudgain.WriteMetadata(ffmpegPath, res, tagMode); err != nil {
			log.Println(err)
		}
	}
}
