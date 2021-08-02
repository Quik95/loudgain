package loudgain

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
)

var Integrated_Loudness_Filter *regexp.Regexp = regexp.MustCompile(`I:\s*(-?\d+\.?\d{1})\sLUFS`)

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

func ParseLounessOutput(data string) error {
	intergratedLoudnessSubgroups := Integrated_Loudness_Filter.FindAllStringSubmatch(data, -1)
	if intergratedLoudnessSubgroups == nil {
		return fmt.Errorf("Can't match integrated loudness\n%s", data)
	}

	integratedLoudness := intergratedLoudnessSubgroups[len(intergratedLoudnessSubgroups)-1][1]

	log.Printf("Integrated loudness of the file is: %s LUFS", integratedLoudness)

	return nil
}
