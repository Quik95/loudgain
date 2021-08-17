package loudgain

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
)

var (
	integratedLoudnessFilter = regexp.MustCompile(`I:\s*(-?\d+\.?\d{1})\sLUFS`)
	loudnessRangeFilter      = regexp.MustCompile(`LRA:\s*(-?\d+\.?\d{1})\sLU`)
	truePeakFilter           = regexp.MustCompile(`Peak:\s*(-?\d+\.?\d{1})\sdBFS`)
)

// NoMatchError indicates that parsing ffmpeg output did not result in obtaining an expected value.
type NoMatchError struct {
	Data string
}

func (e NoMatchError) Error() string {
	return fmt.Sprintf("failed to match: %s", e.Data)
}

// GetFFmpegPath gets the location of an ffmpeg binary in the system.
func GetFFmpegPath() (string, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("failed to located ffmpeg path: %w", err)
	}

	return path, nil
}

// RunLoudnessScan runs ffmpeg ebur128 scan on a given file and captures it's output.
func RunLoudnessScan(filepath string) (string, error) {
	ffmpegPath, err := GetFFmpegPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(
		ffmpegPath,
		"-hide_banner",
		"-i",
		filepath,
		"-filter_complex",
		"ebur128=peak='true':framelog='verbose'",
		"-f",
		"null",
		"-",
	)
	log.Println(cmd.String())

	var output bytes.Buffer
	cmd.Stderr = &output

	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, output.String())
	}

	return output.String(), nil
}

// ParseLoudnessOutput parses ffmpeg ebur128 filter output.
func ParseLoudnessOutput(data string, filepath string) (LoudnessLevel, error) {
	integratedLoudness, err := filterData(data, integratedLoudnessFilter)
	if err != nil {
		return LoudnessLevel{}, fmt.Errorf("failed to match integrated loudness: %w", err)
	}

	loudnessRange, err := filterData(data, loudnessRangeFilter)
	if err != nil {
		return LoudnessLevel{}, fmt.Errorf("failed to match loudness range: %w", err)
	}

	truePeak, err := filterData(data, truePeakFilter)
	if err != nil {
		return LoudnessLevel{}, fmt.Errorf("failed to match true peak: %w", err)
	}

	integratedLoudnessFloat, err := strconv.ParseFloat(integratedLoudness, 64)
	if err != nil {
		return LoudnessLevel{}, fmt.Errorf("failed to convert integrated loudness to float: %w", err)
	}

	loudnessRangeFloat, err := strconv.ParseFloat(loudnessRange, 64)
	if err != nil {
		return LoudnessLevel{}, fmt.Errorf("failed to convert loudness range to float: %w", err)
	}

	truePeakFloat, err := strconv.ParseFloat(truePeak, 64)
	if err != nil {
		return LoudnessLevel{}, fmt.Errorf("failed to convert true peak to float: %w", err)
	}

	ll := LoudnessLevel{
		IntegratedLoudness: integratedLoudnessFloat,
		LoudnessRange:      loudnessRangeFloat,
		TruePeakdB:         truePeakFloat,
		Filepath:           filepath,
	}

	return ll, nil
}

// LoudnessLevel represents the loudness data of a given song.
type LoudnessLevel struct {
	IntegratedLoudness, LoudnessRange, TruePeakdB float64
	Filepath                                      string
}

// String returns a textual representation of this struct.
func (ll LoudnessLevel) String() string {
	return fmt.Sprintf(
		"\nFilepath: %s\n"+
			"Integrated loudness: %f LUFS\n"+
			"Loudness Range: %f LU\n"+
			"True peak in dBFS: %f",
		ll.Filepath, ll.IntegratedLoudness, ll.LoudnessRange, ll.TruePeakdB,
	)
}

func filterData(data string, filter *regexp.Regexp) (string, error) {
	resultWithSubgroups := filter.FindAllStringSubmatch(data, -1)
	if resultWithSubgroups == nil {
		return "", NoMatchError{Data: data}
	}

	return resultWithSubgroups[len(resultWithSubgroups)-1][1], nil
}
