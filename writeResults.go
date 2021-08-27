package loudgain

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// WriteMeta writes converts data provided by the scan argument to their replaygain tag representation.  It then writes these tag to a copy of the original file.
// Next it prepands backup- to the original file and renames the copy to the original.
// When all this finishes successfully, this function removes the original file.
// In case of an error the original file is renamed back to it's initial name, and the copy is deleted.
func WriteMetadata(ffmpegPath string, scan ScanResult, tagMode WriteMode) error {
	tempFile := prependToBase(scan.FilePath, "loudgain-")

	if err := ffmpegWriteMetadata(scan, FFmpegPath, tempFile, tagMode); err != nil {
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

func getFFmpegArgs(metadata ScanResult, swapFile string, mode WriteMode) []string {
	base := []string{
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
	}

	switch mode {
	case WriteRG2:
		base = append(base, []string{
			"-metadata",
			fmt.Sprintf("replaygain_track_gain=%.2f dB", metadata.TrackGain),
			"-metadata",
			fmt.Sprintf("replaygain_track_peak=%.6f", metadata.TrackPeak),
		}...)
	case ExtraTags:
		base = append(base, []string{
			"-metadata",
			fmt.Sprintf("replaygain_track_gain=%.2f dB", metadata.TrackGain),
			"-metadata",
			fmt.Sprintf("replaygain_track_peak=%.6f", metadata.TrackPeak),
			"-metadata",
			fmt.Sprintf("replaygain_reference_loudness=%.2f LUFS", metadata.ReferenceLoudness),
			"-metadata",
			fmt.Sprintf("replaygain_track_range=%.2f dB", metadata.TrackRange),
		}...)
	case ExtraTagsLU:
		base = append(base, []string{
			"-metadata",
			fmt.Sprintf("replaygain_track_gain=%.2f LUFS", metadata.TrackGain.ToLoudnessUnit()),
			"-metadata",
			fmt.Sprintf("replaygain_track_peak=%.6f", metadata.TrackPeak),
			"-metadata",
			fmt.Sprintf("replaygain_reference_loudness=%.2f LUFS", metadata.ReferenceLoudness),
			"-metadata",
			fmt.Sprintf("replaygain_track_range=%.2f LU", metadata.TrackRange.ToLoudnessUnit()),
		}...)
	case DeleteTags, InvalidWriteMode, SkipWritingTags:
	}

	// don't forget to include output path as the last item
	base = append(base, swapFile)

	return base
}

func ffmpegWriteMetadata(metadata ScanResult, ffmpegPath, swapFile string, mode WriteMode) error {
	args := getFFmpegArgs(metadata, swapFile, mode)

	cmd := exec.Command(FFmpegPath, args...)

	var output bytes.Buffer
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed to write metadata: %s\n%w", output.String(), err)
	}

	return nil
}
