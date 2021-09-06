package loudgain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sync"
)

// GetAlbums function reads albums from songs and return a list of unique album names.
func GetAlbums(songs []string) []string {
	var wg sync.WaitGroup
	wg.Add(len(songs))

	reschan := make(chan string, len(songs))
	guard := make(chan struct{}, WorkersLimit)

	for _, song := range songs {
		guard <- struct{}{}
		go func(song string, reschan chan<- string, wg *sync.WaitGroup) {
			album, err := getSongsAlbum(song)
			if err != nil {
				log.Println(err)
			}
			reschan <- album

			<-guard
			wg.Done()
		}(song, reschan, &wg)
	}
	wg.Wait()
	close(reschan)

	uniqueAlbums := map[string]bool{}
	albums := make([]string, 0, len(uniqueAlbums))

	for album := range reschan {
		if _, ok := uniqueAlbums[album]; !ok {
			uniqueAlbums[album] = true
			albums = append(albums, album)
		}
	}

	log.Printf("%#v", albums)
	log.Printf("n songs: %d, n albums: %d", len(songs), len(albums))
	return albums
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
