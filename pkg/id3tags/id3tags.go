package id3tags

import (
	"bytes"
	"fmt"
	"github.com/r3boot/go-musicbot/pkg/config"
	"github.com/sirupsen/logrus"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	ModuleName = "ID3Tags"

	RatingUnknown  = -1
	RatingZero     = 0
	RatingToRemove = 1
	RatingDefault  = 5
	RatingMax      = math.MaxInt32 - 1

	id3Track   = "TRCK"
	id3Artist  = "TPE1"
	id3Title   = "TIT2"
	id3Comment = "COMM"
)

type ID3Tags struct {
	log *logrus.Entry
	cfg *config.MusicBotConfig
}

var (
	reId3v2Fname     = regexp.MustCompile("^id3v2 tag info for (.*)\\:$")
	reId3v2Submitter = regexp.MustCompile("^COMM \\(Comments\\): \\(\\)\\[\\]: (.*)$")

	//COMM (Comments): ()[]: ]V[
)

func NewID3Tags(cfg *config.MusicBotConfig) (*ID3Tags, error) {
	tags := &ID3Tags{
		log: logrus.WithFields(logrus.Fields{
			"caller": ModuleName,
		}),
		cfg: cfg,
	}

	if _, err := os.Stat(cfg.Paths.Id3v2); err != nil {
		return nil, fmt.Errorf("os.Stat: %v", err)
	}

	return tags, nil
}

func (tags *ID3Tags) runId3v2(params []string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("id3v2", params...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("cmd.Start: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		msg := fmt.Sprintf("cmd.Wait: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		return "", fmt.Errorf(msg)
	}

	if stderr.Len() > 0 {
		msg := fmt.Sprintf("failed to run: %v\n", stderr.String())
		return "", fmt.Errorf(msg)
	}

	return stdout.String(), nil
}

func (tags *ID3Tags) readFrame(fname, frame string) (string, error) {
	params := []string{"-l", fname}

	output, err := tags.runId3v2(params)
	if err != nil {
		return "", fmt.Errorf("tags.runId3v2: %v", err)
	}

	regex := fmt.Sprintf("^%s.*: (.*)$", frame)
	reFrame := regexp.MustCompile(regex)
	for _, line := range strings.Split(output, "\n") {
		/* Hack around:
		 * COMM (Comments): (ID3v1 Comment)[XXX]:  blahblah
		 * COMM (Comments): ()[]: blahblah
		 */
		if frame == id3Comment && strings.Contains(line, "ID3v1 Comment") {
			continue
		}

		result := reFrame.FindAllStringSubmatch(line, -1)
		if len(result) == 0 {
			continue
		}

		response := result[0][1]

		return response, nil
	}

	return "", fmt.Errorf("no %s frame found", frame)
}

// Functionality related to submitter
func (tags *ID3Tags) GetSubmitter(fname string) (string, error) {
	submitter, err := tags.readFrame(fname, id3Comment)
	if err != nil {
		return "", fmt.Errorf("tags.readFrame: %v", err)
	}

	return submitter, nil
}

func (tags *ID3Tags) HasSubmitter(fname string) (bool, error) {
	result, err := tags.GetSubmitter(fname)
	if err != nil {
		return false, fmt.Errorf("tags.GetSubmitter: %v", err)
	}

	return len(result) > 0, nil
}

func (tags *ID3Tags) SetSubmitter(fname, nickname string) error {
	hasSubmitter, err := tags.HasSubmitter(fname)
	if err != nil {
		return fmt.Errorf("tags.hasSubmitter: %v", err)
	}

	if hasSubmitter {
		return fmt.Errorf("track already has a submitter")
	}

	nickname = strings.TrimSpace(nickname)

	_, err = tags.runId3v2([]string{"-c", nickname, fname})
	if err != nil {
		return fmt.Errorf("tags.runId3v2: %v", err)
	}

	return nil
}

func (tags *ID3Tags) GetAllSubmitters() map[string]string {
	globPattern := fmt.Sprintf("%s/*.mp3", tags.cfg.Paths.Music)

	files, err := filepath.Glob(globPattern)
	if err != nil {
		tags.log.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to glob files")
		return nil
	}
	params := []string{"-l"}
	params = append(params, files...)

	output, err := tags.runId3v2(params)
	if err != nil {
		tags.log.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Failed to run command")
		return nil
	}

	fname := ""
	submitter := ""
	result := make(map[string]string)

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "id3v2 tag info") {
			result := reId3v2Fname.FindAllStringSubmatch(line, -1)
			if len(result) != 1 {
				tags.log.Warn("No filename found")
				fname = ""
				submitter = ""
				continue
			}
			fname = result[0][1]
		}

		if fname != "" && strings.HasPrefix(line, id3Comment) {
			result := reId3v2Submitter.FindAllStringSubmatch(line, -1)
			if len(result) != 1 {
				tags.log.WithFields(logrus.Fields{
					"line": line,
				}).Warn("No submitter found")
				fname = ""
				submitter = ""
				continue
			}
			submitter = result[0][1]
		}

		if fname != "" && submitter != "" {
			result[fname] = submitter
			fname = ""
			submitter = ""
		}
	}

	return result
}

// Functionality related to ratings
func (tags *ID3Tags) SetRating(fname string, rating int) (int, error) {
	if rating <= RatingZero || rating >= RatingMax {
		return RatingUnknown, fmt.Errorf("rating must be between %d and %d", RatingZero, RatingMax)
	}

	rating_s := strconv.Itoa(rating)

	params := []string{"-T", rating_s, fname}
	_, err := tags.runId3v2(params)
	if err != nil {
		return RatingUnknown, fmt.Errorf("tags.runId3v2: %v", err)
	}

	return rating, nil
}

func (tags *ID3Tags) GetRating(fname string) (int, error) {
	var (
		rating int
		err    error
	)

	rating = RatingUnknown
	result, _ := tags.readFrame(fname, id3Track)

	if result != "" {
		rating, err = strconv.Atoi(result)
		if err != nil {
			return RatingUnknown, fmt.Errorf("strconv.Atoi:%v", err)
		}
	} else {
		rating = RatingUnknown
	}

	return rating, nil
}

func (tags *ID3Tags) DecreaseRating(fname string) (int, error) {
	rating, err := tags.GetRating(fname)
	if err != nil {
		return RatingUnknown, fmt.Errorf("tags.GetRating: %v", err)
	}

	switch rating {
	case RatingUnknown:
		return RatingUnknown, nil
	case RatingZero:
		return RatingZero, nil
	default:
		{
			rating -= 1
			rating, err := tags.SetRating(fname, rating)
			if err != nil {
				return RatingUnknown, fmt.Errorf("tags.SetRating: %v", err)
			}
			return rating, nil
		}
	}

	return RatingUnknown, fmt.Errorf("unknown error")
}

func (tags *ID3Tags) IncreaseRating(fname string) (int, error) {
	rating, err := tags.GetRating(fname)
	if err != nil {
		return RatingUnknown, fmt.Errorf("ID3Tags.IncreaseRating: %v", err)
	}

	switch rating {
	case RatingUnknown:
		return RatingUnknown, nil
	case RatingMax:
		return RatingMax, nil
	default:
		{
			rating += 1
			rating, err := tags.SetRating(fname, rating)
			if err != nil {
				return RatingUnknown, fmt.Errorf("tags.SetRating: %v", err)
			}
			return rating, nil
		}
	}

	return RatingUnknown, fmt.Errorf("unknown error")
}
