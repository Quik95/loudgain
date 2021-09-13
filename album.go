package loudgain

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type songWithAlbum struct {
	Album, Song string
}

// GetProgressBar returns a progressbar with a customized options.
func GetProgressBar(numberOfJobs int) *progressbar.ProgressBar {
	return progressbar.NewOptions(numberOfJobs,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionUseANSICodes(true),
		progressbar.OptionShowCount(),
	)
}

// TimeTrack measures the time between calling it and it's execution.
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func getListOfAlbumSongs(pairs []songWithAlbum) map[string][]string {
	albumsWithSongs := map[string][]string{}

	for _, pair := range pairs {
		if _, ok := albumsWithSongs[pair.Album]; !ok {
			albumsWithSongs[pair.Album] = []string{pair.Song}
		} else {
			albumsWithSongs[pair.Album] = append(albumsWithSongs[pair.Album], pair.Song)
		}
	}

	return albumsWithSongs
}

// GetScannedAlbums function reads albums from songs and return a list of unique album names.
func GetScannedAlbums(songs []string) <-chan ScanResult {
	albumAndSongPairs := make([]songWithAlbum, 0, len(songs))

	for _, song := range songs {
		album, err := getAlbumFromSong(song)
		if err != nil {
			log.Println(err)
		}

		albumAndSongPairs = append(albumAndSongPairs, songWithAlbum{Album: album, Song: song})
	}

	albumsWithSongs := getListOfAlbumSongs(albumAndSongPairs)

	log.Printf("Found %d albums in %d songs.", len(albumsWithSongs), len(songs))

	pg := GetProgressBar(len(albumsWithSongs))
	pg.Describe("Scanning albums")

	var wg sync.WaitGroup

	wg.Add(len(albumsWithSongs))

	numberOfSongs := 0
	for _, songs := range albumsWithSongs {
		numberOfSongs += len(songs)
	}

	reschan := make(chan ScanResult, numberOfSongs)
	guard := make(chan struct{}, WorkersLimit)

	scan := func(songs []string) {
		res := scanAlbum(songs)
		for _, x := range res {
			reschan <- x
		}

		wg.Done()
		pg.Add(1)
		<-guard
	}

	for _, songs := range albumsWithSongs {
		guard <- struct{}{}

		go scan(songs)
	}

	wg.Wait()
	close(reschan)

	return reschan
}

func scanAlbum(songs []string) []ScanResult {
	// there is nothing to join if album has only one song
	if len(songs) == 1 {
		return []ScanResult{ScanFile(songs[0])}
	}

	results := make([]ScanResult, 0, len(songs))

	filename, err := combineIntoOneFile(songs)
	if err != nil {
		log.Printf("failed to scan songs: %#v", songs)
		log.Println(err)

		return nil
	}
	defer os.Remove(filename)

	tempScanResult := ScanFile(filename)

	for _, song := range songs {
		results = append(results, ScanResult{
			FilePath:          song,
			TrackGain:         tempScanResult.TrackGain,
			TrackRange:        tempScanResult.TrackRange,
			TrackPeak:         tempScanResult.TrackPeak,
			Loudness:          tempScanResult.Loudness,
			ReferenceLoudness: ReferenceLoudness,
		})
	}

	return results
}

func getHashOfStrings(in []string) string {
	hash := sha1.New()

	for _, str := range in {
		hash.Write([]byte(str))
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func checkSameExtension(songs []string) bool {
	if len(songs) < 1 {
		return false
	}

	res := true
	last := filepath.Ext(songs[0])

	for _, song := range songs {
		ext := filepath.Ext(song)
		if ext != last {
			res = false

			break
		}
	}

	return res
}

func combineIntoOneFile(songs []string) (string, error) {
	if sameExt := checkSameExtension(songs); !sameExt {
		return "", errors.New("calculating the album replaygain across multiple filetypes is not supported")
	}

	ffmpegConcatFile, err := writeFFmpegConcatInput(songs)
	if err != nil {
		return "", fmt.Errorf("failed to combine into one file for scanning album gain: %w", err)
	}
	defer os.Remove(ffmpegConcatFile.Name())

	concatSongFilepath := filepath.Join(
		os.TempDir(),
		getHashOfStrings(songs),
	) + filepath.Ext(songs[0])

	cmd := exec.Command(
		FFmpegPath,
		"-hide_banner",
		"-y",
		"-f",
		"concat",
		"-safe",
		"0",
		"-i",
		ffmpegConcatFile.Name(),
		"-c",
		"copy",
		concatSongFilepath,
	)

	_, err = cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return "", fmt.Errorf("failed to concatenate songs: %s", exitError.Stderr)
		}

		return "", fmt.Errorf("failed to concatenate songs: %w", err)
	}

	return concatSongFilepath, nil
}

func escapeQuotes(filename string) string {
	escaped := strings.ReplaceAll(filename, `'`, `'\''`)

	return escaped
}

func writeFFmpegConcatInput(songs []string) (*os.File, error) {
	file, err := os.CreateTemp("", getHashOfStrings(songs))
	if err != nil {
		return nil, fmt.Errorf("failed to create a ffmpeg concat input file: %w", err)
	}

	for _, song := range songs {
		str := fmt.Sprintf("file '%s'\n", escapeQuotes(song))
		file.WriteString(str)
	}

	return file, nil
}

func getAlbumFromSong(song string) (string, error) {
	cmd := exec.Command(
		FFprobePath,
		"-show_format",
		"-print_format",
		"json",
		song,
	)

	output, err := cmd.Output()
	if err != nil {
		var execError *exec.ExitError
		if errors.As(err, &execError) {
			return "", fmt.Errorf("failed to probe %s: %s", song, execError.Stderr)
		}

		return "", fmt.Errorf("failed to probe %s", song)
	}

	type ffprobeResult struct {
		Format struct {
			Tags struct {
				Album string `json:"album"`
			} `json:"tags"`
		} `json:"format"`
	}

	var res ffprobeResult

	if err := json.Unmarshal(output, &res); err != nil {
		return "", fmt.Errorf("failed to decode json for song %s: %w", song, err)
	}

	return res.Format.Tags.Album, nil
}
