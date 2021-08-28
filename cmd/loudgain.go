package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"runtime"

	"github.com/Quik95/loudgain"
)

var (
	flagPeakLimit, flagPregain float64
	noClip, quiet              bool
	tagMode                    string
	numberOfWorkers            int
)

func init() {
	flag.Float64Var(&flagPeakLimit, "maxtpl", -1.0, "Maximal true peak level in dBTP")
	flag.Float64Var(&flagPregain, "pregain", 0.0, "Apply n dB/LU pre-gain value")
	flag.IntVar(&numberOfWorkers, "workers", runtime.NumCPU(), "Number of workers scanning songs in parallel.")
	flag.BoolVar(&quiet, "quiet", false, "Supress output.")
	flag.BoolVar(&noClip, "noclip", false, "Lower track gain to avoid clipping.")
	flag.StringVar(&tagMode, "tagmode", "s",
		"--tagmode=d Delete ReplayGain tags from files. (uninmplemented)\n"+
			"--tagmode=i Write Replaygain 2.0 tags to files.\n"+
			"--tagmode=e like --tagmode=i, plus extra tags (reference, ranges).\n"+
			"--tagmode=l like --tagmode=e, but LU units instead of dB.\n"+
			"--tagmode=s Don't write Replaygain tags.")

	flag.Parse()
}

func checkExitCondition(tagMode loudgain.WriteMode) error {
	if flag.NArg() == 0 {
		return errors.New("No files to process. Exitting...")
	}

	if tagMode == loudgain.InvalidWriteMode {
		return errors.New("Invalid write mode. Exitting...")
	}

	return nil
}

func setGlobals() {
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
}

func main() {
	if err := checkExitCondition(loudgain.TagMode); err != nil {
		log.Fatal(err)
	}

	setGlobals()

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
		res := <-results
		if !quiet {
			fmt.Println(res)
		}
	}

	close(jobs)
}

func worker(jobs <-chan string, results chan<- loudgain.ScanResult) {
	for job := range jobs {
		res := loudgain.ScanFile(job)
		results <- res
	}
}
