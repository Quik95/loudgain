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
	flagPeakLimit, flagPregain float64
	noClip, quiet, album       bool
	tagMode                    string
	numberOfWorkers            int
)

func init() {
	flag.Float64Var(&flagPeakLimit, "maxtpl", -1.0, "Maximal true peak level in dBTP")
	flag.Float64Var(&flagPregain, "pregain", 0.0, "Apply n dB/LU pre-gain value")
	flag.IntVar(&numberOfWorkers, "workers", runtime.NumCPU(), "Number of workers scanning songs in parallel.")
	flag.BoolVar(&quiet, "quiet", false, "Supress output.")
	flag.BoolVar(&noClip, "noclip", false, "Lower track gain to avoid clipping.")
	flag.BoolVar(&album, "album", true, "Also calculate replaygain values for album.")
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
		} else {
			if err := loudgain.CheckExtension(path); err == nil {
				songs = append(songs, path)
			}
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

	var wg sync.WaitGroup

	guard := make(chan struct{}, loudgain.WorkersLimit)

	if album {
		scannedSongs := loudgain.GetScannedAlbums(songs)

		wg.Add(len(scannedSongs))

		for song := range scannedSongs {
			go writeToSong(song, true, guard, &wg)
		}

		wg.Wait()

		return
	}

	wg.Add(len(songs))

	progressBar := loudgain.GetProgressBar(len(songs))
	progressBar.Describe("Scanning songs")

	results := make(chan loudgain.ScanResult, len(songs))

	scanSong := func(song string) {
		guard <- struct{}{}
		results <- loudgain.ScanFile(song)

		wg.Done()
		progressBar.Add(1)
		<-guard
	}

	for _, song := range songs {
		go scanSong(song)
	}
	wg.Wait()
	close(results)

	wg.Add(len(results))
	for res := range results {
		go writeToSong(res, false, guard, &wg)
	}
	wg.Wait()
}

func writeToSong(sr loudgain.ScanResult, album bool, guard chan struct{}, wg *sync.WaitGroup) {
	guard <- struct{}{}
	if err := loudgain.WriteMetadata(sr, album); err != nil {
		log.Println(err)
	}
	if !quiet {
		fmt.Println(sr)
	}

	wg.Done()
	<-guard
}
