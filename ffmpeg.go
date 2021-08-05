package loudgain

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os/exec"
	"regexp"
	"strconv"
)

var (
	integratedLoudnessFilter *regexp.Regexp = regexp.MustCompile(`I:\s*(-?\d+\.?\d{1})\sLUFS`)
	loudnessRangeFilter      *regexp.Regexp = regexp.MustCompile(`LRA:\s*(-?\d+\.?\d{1})\sLU`)
	truePeakFilter           *regexp.Regexp = regexp.MustCompile(`Peak:\s*(-?\d+\.?\d{1})\sdBFS`)
)

// GetFFmpegPath gets the location of an ffmpeg binary in the system.
func GetFFmpegPath() (string, error) {
	return exec.LookPath("ffmpeg")
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
		return "", fmt.Errorf(output.String())
	}

	return output.String(), nil
}

// ParseLoudnessOutput parses ffmpeg ebur128 filter output.
func ParseLounessOutput(data string, filepath string) (LoudnessLevel, error) {
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
		return LoudnessLevel{}, err
	}

	loudnessRangeFloat, err := strconv.ParseFloat(loudnessRange, 64)
	if err != nil {
		return LoudnessLevel{}, err
	}

	truePeakFloat, err := strconv.ParseFloat(truePeak, 64)
	if err != nil {
		return LoudnessLevel{}, err
	}

	ll := LoudnessLevel{
		IntegratedLoudness: integratedLoudnessFloat,
		LoudnessRange:      loudnessRangeFloat,
		TruePeak:           decibelToLinear(truePeakFloat),
		Filepath:           filepath,
	}

	return ll, nil
}

// LoudnessLevel represents the loudness data of a given song.
type LoudnessLevel struct {
	IntegratedLoudness, LoudnessRange, TruePeak float64
	Filepath                                    string
}

// String returns a textual representation of this struct.
func (ll LoudnessLevel) String() string {
	return fmt.Sprintf(
		"\nFilepath: %s\n"+
			"Integrated loudness: %f LUFS\n"+
			"Loudness Range: %f LU\n"+
			"True peak: %f",
		ll.Filepath, ll.IntegratedLoudness, ll.LoudnessRange, ll.TruePeak,
	)
}

func filterData(data string, filter *regexp.Regexp) (string, error) {
	resultWithSubgroups := filter.FindAllStringSubmatch(data, -1)
	if resultWithSubgroups == nil {
		return "", fmt.Errorf("failed to match\n%s", data)
	}

	return resultWithSubgroups[len(resultWithSubgroups)-1][1], nil
}

func decibelToLinear(in float64) float64 {
	return math.Pow(10.0, in/20)
}
