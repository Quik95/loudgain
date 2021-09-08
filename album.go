package loudgain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type songWithAlbum struct {
	Album, Song string
}

func GetProgressBar(numberOfJobs int) *progressbar.ProgressBar {
	return progressbar.NewOptions(numberOfJobs,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionUseANSICodes(true),
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

	return albumsWithSongs
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
