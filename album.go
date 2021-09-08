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

// GetAlbums function reads albums from songs and return a list of unique album names.
func GetAlbums(songs []string) map[string][]string {
	var wg sync.WaitGroup

	wg.Add(len(songs))

	reschan := make(chan songWithAlbum, len(songs))
	guard := make(chan struct{}, WorkersLimit)

	progressBar := GetProgressBar(len(songs))
	progressBar.Describe("Getting albums from songs")

	runGetSongsInParallel := func(song string, reschan chan<- songWithAlbum, wg *sync.WaitGroup) {
		album, err := getSongsAlbum(song)
		if err != nil {
			log.Println(err)
		}
		reschan <- songWithAlbum{Album: album, Song: song}

		<-guard
		wg.Done()
		progressBar.Add(1)
	}

	for _, song := range songs {
		guard <- struct{}{}

		go runGetSongsInParallel(song, reschan, &wg)
	}

	wg.Wait()
	close(reschan)

	albumsWithSongs := map[string][]string{}

	for pair := range reschan {
		if _, ok := albumsWithSongs[pair.Album]; !ok {
			albumsWithSongs[pair.Album] = []string{pair.Song}
		} else {
			albumsWithSongs[pair.Album] = append(albumsWithSongs[pair.Album], pair.Song)
		}
	}

	log.Printf("n songs: %d, n albums: %d", len(songs), len(albumsWithSongs))

	filename, err := combineIntoOneFile(albumsWithSongs["Hot Fuss"])
	if err != nil {
		log.Fatalln(err)
	}
	defer os.Remove(filename)

	sr := ScanFile(filename)
	fmt.Println(sr)

	return albumsWithSongs
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

	ffmpegErrorOutput, err := cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return "", fmt.Errorf("failed to concatenate songs: %s", ffmpegErrorOutput)
		}

		return "", fmt.Errorf("failed to concatenate songs: %w", err)
	}

	return concatSongFilepath, nil
}

func writeFFmpegConcatInput(songs []string) (*os.File, error) {
	file, err := os.CreateTemp("", getHashOfStrings(songs))
	if err != nil {
		return nil, fmt.Errorf("failed to create a ffmpeg concat input file: %w", err)
	}

	for _, song := range songs {
		str := fmt.Sprintf("file '%s'\n", song)
		file.WriteString(str)
	}

	return file, err
}

func getSongsAlbum(song string) (string, error) {
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
