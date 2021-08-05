package main

import (
	"log"
	"os"

	"github.com/Quik95/loudgain"
)

const filepath = "/tmp/song.mp3"

func main() {
	songs := os.Args[1:]
	log.Println(songs)

	ffmpegPath, err := loudgain.GetFFmpegPath()
	if err != nil {
		log.Fatalln("ffmpeg not found in path")
	}

	log.Printf("ffmpeg is located at: %s", ffmpegPath)

	loudness, err := loudgain.RunLoudnessScan(filepath)
	if err != nil {
		log.Fatalf("failed to get loudness ratings: %s", err)
	}

	ll, err := loudgain.ParseLounessOutput(loudness, filepath)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(ll)
}
