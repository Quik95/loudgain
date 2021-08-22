package loudgain

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func WriteMetadata(ffmpegPath string, scan ScanResult) error {
	tempFile, err := createSwapFile(scan.FilePath)
	if err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	if err = ffmpegWriteMetadata(scan, ffmpegPath, tempFile); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	defer os.Remove(tempFile)

	log.Printf("outfile: %s", tempFile)

	if err := swapFiles(scan.FilePath, tempFile); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func swapFiles(original, swap string) error {
	directory := filepath.Dir(original)
	swapFileSameDirectoryAsOriginal := filepath.Join(directory, filepath.Base(swap))

	if err := copyFile(swap, swapFileSameDirectoryAsOriginal); err != nil {
		return fmt.Errorf("failed to swap files: %w", err)
	}

	backupName := filepath.Join(filepath.Dir(original), "loudgain-"+filepath.Base(original))
	if err := os.Rename(original, backupName); err != nil {
		return fmt.Errorf("failed to swap files: %w", err)
	}

	if err := os.Rename(swapFileSameDirectoryAsOriginal, original); err != nil {
		return fmt.Errorf("failed to swap files: %w", err)
	}

	if err := os.Remove(backupName); err != nil {
		return fmt.Errorf("failed to remove the backup file: %w", err)
	}

	return nil
}

func copyFile(input, destination string) error {
	in, err := ioutil.ReadFile(input)
	if err != nil {
		return fmt.Errorf("failed to read from a file: %w", err)
	}

	if err := ioutil.WriteFile(destination, in, 0644); err != nil {
		return fmt.Errorf("failed to open a file for writing: %w", err)
	}

	return nil
}

func ffmpegWriteMetadata(metadata ScanResult, ffmpegPath, swapFile string) error {
	args := []string{
		"-hide_banner",
		"-i",
		metadata.FilePath,
		"-map",
		"0",
		"-y",
		"-codec",
		"copy",
		"-write_id3v2",
		"1",
		"-metadata",
		fmt.Sprintf("replaygain_track_gain=%.2f dB", metadata.TrackGain),
		"-metadata",
		fmt.Sprintf("replaygain_track_peak=%.6f", metadata.TrackPeak),
		"-metadata",
		fmt.Sprintf("replaygain_reference_loudness=%.2f LUFS", metadata.ReferenceLoudness),
		"-metadata",
		fmt.Sprintf("replaygain_track_range=%.2f dB", metadata.TrackRange),
		swapFile,
	}

	cmd := exec.Command(ffmpegPath, args...)
	log.Println(cmd.String())

	var output bytes.Buffer
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed to write metadata: %s\n%w", output.String(), err)
	}

	return nil
}

func getExtension(filename string) (string, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return "", errors.New("invalid extension")
	}

	return ext, nil
}

func createSwapFile(filename string) (string, error) {
	ext, err := getExtension(filename)
	if err != nil {
		return "", fmt.Errorf("failed get swap file name: %w", err)
	}

	songName := strings.TrimSuffix(filepath.Base(filename), ext)

	h := md5.New()
	io.WriteString(h, songName)

	// swapFile := filepath.Join(os.TempDir(), hex.EncodeToString(h.Sum(nil))+ ext)
	swapFile := filepath.Join("/tmp/loudgain", hex.EncodeToString(h.Sum(nil))+ext)
	return swapFile, nil
}
