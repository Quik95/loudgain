package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"

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
		log.Fatalln("No files to process. Exitting...")
	}

	if tagMode == loudgain.InvalidWriteMode {
		log.Fatalln("Invalid write mode. Exitting...")
	}
}

func main() {
	var numberOfWorkers = runtime.NumCPU()

	ffmpegPath, err := loudgain.GetFFmpegPath()
	if err != nil {
		log.Fatalln("an ffmpeg binary not found in the path")
	}

	loudgain.ReferenceLoudness = -18
	loudgain.TrackPeakLimit = loudgain.Decibel(flagPeakLimit)
	loudgain.Pregain = loudgain.LoudnessUnit(flagPregain)
	loudgain.TagMode = loudgain.StringToWriteMode(tagMode)
	loudgain.NoClip = noClip
	loudgain.FFmpegPath = ffmpegPath

	checkExitCondition(loudgain.TagMode)

	songs := flag.Args()

	numberOfJobs := len(songs)
	jobs := make(chan string, numberOfJobs)
	results := make(chan loudgain.ScanResult, numberOfJobs)

	for i := 0; i < numberOfWorkers; i++ {
		go worker(jobs, results)
	}

	for _, song := range songs {
		jobs <- song
	}

	for i := 0; i < numberOfJobs; i++ {
		fmt.Println(<-results)
	}

	close(jobs)
}

func worker(jobs <-chan string, results chan<- loudgain.ScanResult) {
	for job := range jobs {
		res := loudgain.ScanFile(job)
		results <- res
	}
}
