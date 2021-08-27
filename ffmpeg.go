package loudgain

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

var (
	integratedLoudnessFilter = regexp.MustCompile(`I:\s*(-?\d+\.?\d{1})\sLUFS`)
	loudnessRangeFilter      = regexp.MustCompile(`LRA:\s*(-?\d+\.?\d{1})\sLU`)
	truePeakFilter           = regexp.MustCompile(`Peak:\s*(-?\d+\.?\d{1})\sdBFS`)
)

// Decibel type describes loudness in decibels.
type Decibel float64

func (d Decibel) String() string {
	return fmt.Sprintf("%.2f dB", d)
}

// LinearLoudness type describes loudness as a linear scale ranging from 0 to 1.
type LinearLoudness float64

// LoudnessUnit type describes loudness in the LU or LUFS unit.
type LoudnessUnit float64

func (l LoudnessUnit) String() string {
	return fmt.Sprintf("%.2f LU", l)
}

// ErrNoMatch indicates that parsing ffmpeg output did not result in obtaining an expected value.
type ErrNoMatch struct {
	Data string
}

func (e ErrNoMatch) Error() string {
	return fmt.Sprintf("failed to match: %s", e.Data)
}

func checkFilename(filename string) error {
	_, err := os.Stat(filename)

	return err
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
	if err := checkFilename(filepath); err != nil {
		return "", err
	}

	cmd := exec.Command(
		FFmpegPath,
		"-hide_banner",
		"-i",
		filepath,
		"-filter_complex",
		"ebur128=peak='true':framelog='verbose'",
		"-f",
		"null",
		"-",
	)

	var output bytes.Buffer
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
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
		IntegratedLoudness: LoudnessUnit(integratedLoudnessFloat),
		LoudnessRange:      LoudnessUnit(loudnessRangeFloat),
		TruePeakdB:         Decibel(truePeakFloat),
		Filepath:           filepath,
	}

	return ll, nil
}

// LoudnessLevel represents the loudness data of a given song.
type LoudnessLevel struct {
	IntegratedLoudness, LoudnessRange LoudnessUnit
	TruePeakdB                        Decibel
	Filepath                          string
}

// String returns a textual representation of this struct.
func (ll LoudnessLevel) String() string {
	return fmt.Sprintf(
		"\nFilepath: %s\n"+
			"Integrated loudness: %f LUFS\n"+
			"Loudness Range: %s\n"+
			"True peak in dBFS: %f",
		ll.Filepath, ll.IntegratedLoudness, ll.LoudnessRange, ll.TruePeakdB,
	)
}

func filterData(data string, filter *regexp.Regexp) (string, error) {
	resultWithSubgroups := filter.FindAllStringSubmatch(data, -1)
	if resultWithSubgroups == nil {
		return "", ErrNoMatch{Data: data}
	}

	return resultWithSubgroups[len(resultWithSubgroups)-1][1], nil
}
