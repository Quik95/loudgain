package main

import (
	"log"
	"os"

	"github.com/Quik95/loudgain"
)

func main() {
	songs := os.Args[1:]
	log.Println(songs)

	ffmpegPath, err := loudgain.GetFFmpegPath()
	if err != nil {
		log.Fatalln("ffmpeg not found in path")
	}

	log.Printf("ffmpeg is located at: %s", ffmpegPath)

	loudness, err := loudgain.RunLoudnessScan("/tmp/song.mp3")
	if err != nil {
		log.Fatalf("failed to get loudness ratings: %s", err)
	}

	// log.Println(loudness)

	loudgain.ParseLounessOutput(loudness)
}
