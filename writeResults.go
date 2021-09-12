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
func WriteMetadata(scan ScanResult, album bool) error {
	tempFile := prependToBase(scan.FilePath, "loudgain-")

	if err := ffmpegWriteMetadata(scan, tempFile, album); err != nil {
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

func getTagsAlbum(metadata ScanResult) []string {
	switch TagMode {
	case WriteRG2:
		return []string{
			"-metadata",
			fmt.Sprintf("replaygain_album_gain=%.2f dB", metadata.TrackGain),
			"-metadata",
			fmt.Sprintf("replaygain_album_peak=%.6f", metadata.TrackPeak),
		}
	case ExtraTags:
		return []string{
			"-metadata",
			fmt.Sprintf("replaygain_album_gain=%.2f dB", metadata.TrackGain),
			"-metadata",
			fmt.Sprintf("replaygain_album_peak=%.6f", metadata.TrackPeak),
			"-metadata",
			fmt.Sprintf("replaygain_reference_loudness=%.2f LUFS", metadata.ReferenceLoudness),
			"-metadata",
			fmt.Sprintf("replaygain_album_range=%.2f dB", metadata.TrackRange),
		}
	case ExtraTagsLU:
		return []string{
			"-metadata",
			fmt.Sprintf("replaygain_album_gain=%.2f LUFS", metadata.TrackGain.ToLoudnessUnit()),
			"-metadata",
			fmt.Sprintf("replaygain_album_peak=%.6f", metadata.TrackPeak),
			"-metadata",
			fmt.Sprintf("replaygain_reference_loudness=%.2f LUFS", metadata.ReferenceLoudness),
			"-metadata",
			fmt.Sprintf("replaygain_album_range=%.2f LU", metadata.TrackRange.ToLoudnessUnit()),
		}
	case DeleteTags, InvalidWriteMode, SkipWritingTags:
		return []string{}
	}

	return []string{}
}

func getTagsTrack(metadata ScanResult) []string {
	switch TagMode {
	case WriteRG2:
		return []string{
			"-metadata",
			fmt.Sprintf("replaygain_track_gain=%.2f dB", metadata.TrackGain),
			"-metadata",
			fmt.Sprintf("replaygain_track_peak=%.6f", metadata.TrackPeak),
		}
	case ExtraTags:
		return []string{
			"-metadata",
			fmt.Sprintf("replaygain_track_gain=%.2f dB", metadata.TrackGain),
			"-metadata",
			fmt.Sprintf("replaygain_track_peak=%.6f", metadata.TrackPeak),
			"-metadata",
			fmt.Sprintf("replaygain_reference_loudness=%.2f LUFS", metadata.ReferenceLoudness),
			"-metadata",
			fmt.Sprintf("replaygain_track_range=%.2f dB", metadata.TrackRange),
		}
	case ExtraTagsLU:
		return []string{
			"-metadata",
			fmt.Sprintf("replaygain_track_gain=%.2f LUFS", metadata.TrackGain.ToLoudnessUnit()),
			"-metadata",
			fmt.Sprintf("replaygain_track_peak=%.6f", metadata.TrackPeak),
			"-metadata",
			fmt.Sprintf("replaygain_reference_loudness=%.2f LUFS", metadata.ReferenceLoudness),
			"-metadata",
			fmt.Sprintf("replaygain_track_range=%.2f LU", metadata.TrackRange.ToLoudnessUnit()),
		}
	case DeleteTags, InvalidWriteMode, SkipWritingTags:
		return []string{}
	}

	return []string{}
}

func getFFmpegArgs(metadata ScanResult, swapFile string, album bool) []string {
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

	var formattedTags []string
	if album {
		formattedTags = getTagsAlbum(metadata)
	} else {
		formattedTags = getTagsTrack(metadata)
	}

	base = append(base, formattedTags...)

	// don't forget to include output path as the last item
	base = append(base, swapFile)

	return base
}

func ffmpegWriteMetadata(metadata ScanResult, swapFile string, album bool) error {
	args := getFFmpegArgs(metadata, swapFile, album)

	cmd := exec.Command(FFmpegPath, args...)

	var output bytes.Buffer
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed to write metadata: %s\n%w", output.String(), err)
	}

	return nil
}
