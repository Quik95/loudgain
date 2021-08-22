package loudgain

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// WriteMeta writes converts data provided by the scan argument to their replaygain tag representation.  It then writes these tag to a copy of the original file.
// Next it prepands backup- to the original file and renames the copy to the original.
// When all this finishes successfully, this function removes the original file.
// In case of an error the original file is renamed back to it's initial name, and the copy is deleted.
func WriteMetadata(ffmpegPath string, scan ScanResult) error {
	tempFile := prependToBase(scan.FilePath, "loudgain-")

	if err := ffmpegWriteMetadata(scan, ffmpegPath, tempFile); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}
	if err := swapFiles(scan.FilePath, tempFile); err != nil {
		if err2 := os.Remove(tempFile); err2 != nil {
			return fmt.Errorf("failed to remove tempFile after an error: %w\n%v", err, err2)
		}
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

func prependToBase(in, prefix string) string {
	directory := filepath.Dir(in)
	base := filepath.Base(in)

	return filepath.Join(directory, prefix+base)
}

func swapFiles(original, swap string) error {
	backupName := prependToBase(original, "backup-")
	if err := os.Rename(original, backupName); err != nil {
		return fmt.Errorf("failed to swap files: %w", err)
	}
	if err := os.Rename(swap, original); err != nil {
		if err2 := os.Rename(backupName, original); err2 != nil {
			return fmt.Errorf("failed to recover from the error: %w\n%v", err, err2)
		}
		return fmt.Errorf("failed to swap files: %w", err)
	}
	if err := os.Remove(backupName); err != nil {
		return fmt.Errorf("failed to remove the backup file: %w", err)
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
