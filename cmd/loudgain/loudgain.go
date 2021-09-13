package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/Quik95/loudgain"
)

var (
	flagPeakLimit, flagPregain  float64
	noClip, quiet, album, track bool
	tagMode                     string
	numberOfWorkers             int
)

func init() {
	flag.Float64Var(&flagPeakLimit, "maxtpl", -1.0, "Maximal true peak level in dBTP")
	flag.Float64Var(&flagPregain, "pregain", 0.0, "Apply n dB/LU pre-gain value")
	flag.IntVar(&numberOfWorkers, "workers", runtime.NumCPU(), "Number of workers scanning songs in parallel.")
	flag.BoolVar(&quiet, "quiet", false, "Suppress output.")
	flag.BoolVar(&noClip, "noclip", false, "Lower track gain to avoid clipping.")
	flag.BoolVar(&album, "album", false, "Also calculate replaygain values for album.")
	flag.BoolVar(&track, "track", true, "Calculate replaygain values for tracks.")
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

func checkIfPathIsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func setGlobals() error {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return errors.New("an ffmpeg binary not found in the path")
	}

	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return errors.New("an ffprobe binary not found in the path")
	}

	loudgain.ReferenceLoudness = -18
	loudgain.TrackPeakLimit = loudgain.Decibel(flagPeakLimit)
	loudgain.Pregain = loudgain.LoudnessUnit(flagPregain)
	loudgain.TagMode = loudgain.StringToWriteMode(tagMode)
	loudgain.NoClip = noClip
	loudgain.FFmpegPath = ffmpegPath
	loudgain.FFprobePath = ffprobePath
	loudgain.WorkersLimit = numberOfWorkers

	return nil
}

func getSongsFromDirectory(dir string) (songs []string) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range dirEntries {
		if entry.IsDir() {
			songs = append(songs, getSongsFromDirectory(filepath.Join(dir, entry.Name()))...)
		} else {
			songs = append(songs, filepath.Join(dir, entry.Name()))
		}
	}

	return
}

func expandSongs(paths []string) (songs []string) {
	for _, path := range paths {
		if checkIfPathIsDirectory(path) {
			songsInDirectory := getSongsFromDirectory(path)
			for _, song := range songsInDirectory {
				if err := loudgain.CheckExtension(song); err == nil {
					songs = append(songs, song)
				}
			}
		} else if err := loudgain.CheckExtension(path); err == nil {
			songs = append(songs, path)
		}
	}

	return
}

func main() {
	if err := checkExitCondition(loudgain.TagMode); err != nil {
		log.Fatal(err)
	}

	if err := setGlobals(); err != nil {
		log.Fatalln(err)
	}

	songs := expandSongs(flag.Args())

	if album {
		scannedSongs := loudgain.GetScannedAlbums(songs)

		writeToSongs(scannedSongs, true)
	}

	if track {
		scannedSongs := loudgain.GetScannedSongs(songs)

		writeToSongs(scannedSongs, false)
	}
}

func writeToSongs(scanchan <-chan loudgain.ScanResult, album bool) {
	var wg sync.WaitGroup

	wg.Add(len(scanchan))

	guard := make(chan struct{}, loudgain.WorkersLimit)

	write := func(scan loudgain.ScanResult) {
		if err := loudgain.WriteMetadata(scan, album); err != nil {
			log.Println(err)
		}

		if !quiet {
			fmt.Println(scan)
		}

		<-guard
		wg.Done()
	}

	for scan := range scanchan {
		guard <- struct{}{}

		go write(scan)
	}

	wg.Wait()
}
