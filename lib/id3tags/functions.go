package id3tags

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"path"
)

func (i *ID3Tags) expandFullPath(fname string) (string, error) {
	var err error

	fullPath := fname
	if !strings.HasPrefix(fname, "/") {
		fullPath, err = filepath.Abs(i.BaseDir + "/" + fname)
		if err != nil {
			return "", fmt.Errorf("i.expandFullPath filepath.Abs: %v", err)
		}
	}

	return fullPath, nil
}

func (i *ID3Tags) runId3v2(params []string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("id3v2", params...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr


	log.Debugf("Running %s %s\n", "id3v2", strings.Join(params, " "))

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("ID3Tags.runId3v2 cmd.Start: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		msg := fmt.Sprintf("ID3Tags.runId3v2 cmd.Wait: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		return "", fmt.Errorf(msg)
	}

	if stderr.Len() > 0 {
		msg := fmt.Sprintf("ID3Tags.runId3v2: failed to run: %v\n", stderr.String())
		return "", fmt.Errorf(msg)
	}

	return stdout.String(), nil
}

func (i *ID3Tags) ReadFrame(fname, frame string) (string, error) {
	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.ReadField: %v", err)
	}

	params := []string{"-l", fullPath}

	output, err := i.runId3v2(params)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.ReadField: %v", err)
	}

	regex := fmt.Sprintf("^%s.*: (.*)$", frame)
	reFrame := regexp.MustCompile(regex)
	for _, line := range strings.Split(output, "\n") {
		result := reFrame.FindAllStringSubmatch(line, -1)
		if len(result) == 0 {
			continue
		}

		response := result[0][1]

		return response, nil
	}

	return "", fmt.Errorf("ID3Tags.ReadField: no %s frame found", frame)
}

func (i *ID3Tags) SetRating(fname string, rating int) (int, error) {
	if rating <= RATING_ZERO || rating >= RATING_MAX {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.SetRating: rating must be between %d and %d", RATING_ZERO, RATING_MAX)
	}

	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.SetRating: %v", err)
	}

	rating_s := strconv.Itoa(rating)

	params := []string{"-T", rating_s, fullPath}
	_, err = i.runId3v2(params)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.SetRating: %v", err)
	}

	return rating, nil
}

func (i *ID3Tags) GetRating(fname string) (int, error) {
	var rating int

	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.GetRating: %v", err)
	}

	rating = RATING_UNKNOWN
	result, _ := i.ReadFrame(fullPath, TRACK)

	if result != "" {
		rating, err = strconv.Atoi(result)
		if err != nil {
			return RATING_UNKNOWN, fmt.Errorf("ID3Tags.GetRating strconv.Atoi: %v", err)
		}
	} else {
		rating = RATING_UNKNOWN
	}

	return rating, nil
}

func (i *ID3Tags) DecreaseRating(name string) (int, error) {
	fullPath, err := i.expandFullPath(name)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.DecreaseRating: %v", err)
	}

	rating, err := i.GetRating(fullPath)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.DecreaseRating: %v", err)
	}

	switch rating {
	case RATING_UNKNOWN:
		return RATING_UNKNOWN, nil
	case RATING_ZERO:
		return RATING_ZERO, nil
	default:
		{
			rating -= 1
			rating, err := i.SetRating(fullPath, rating)
			if err != nil {
				return RATING_UNKNOWN, fmt.Errorf("ID3Tags.DecreaseRating: %v", err)
			}
			return rating, nil
		}
	}

	return RATING_UNKNOWN, fmt.Errorf("ID3Tags.DecreaseRating: unknown error")
}

func (i *ID3Tags) IncreaseRating(name string) (int, error) {
	fullPath, err := i.expandFullPath(name)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.IncreaseRating: %v", err)
	}

	rating, err := i.GetRating(fullPath)
	if err != nil {
		return RATING_UNKNOWN, fmt.Errorf("ID3Tags.IncreaseRating: %v", err)
	}

	switch rating {
	case RATING_UNKNOWN:
		return RATING_UNKNOWN, nil
	case RATING_MAX:
		return RATING_MAX, nil
	default:
		{
			rating += 1
			rating, err := i.SetRating(name, rating)
			if err != nil {
				return RATING_UNKNOWN, fmt.Errorf("ID3Tags.IncreaseRating: %v", err)
			}
			return rating, nil
		}
	}

	return RATING_UNKNOWN, fmt.Errorf("ID3Tags.IncreaseRating: unknown error")
}

func (i *ID3Tags) RemoveFile(fname string) error {
	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return fmt.Errorf("ID3Tags.RemoveFile: %v", err)
	}

	_, err = os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("ID3Tags.RemoveFile os.Stat: %v", err)
	}

	err = os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("ID3Tags.RemoveFile os.Remove: %v", err)
	}

	return nil
}

func (i *ID3Tags) GetAllFiles() ([]string, error) {
	files, err := ioutil.ReadDir(i.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("ID3Tags.GetAllFiles ioutil.ReadDir: %v", err)
	}

	tmpList := make([]string, len(files))
	totItems := 0
	for _, fs := range files {
		if fs.Name() == "" {
			continue
		}
		if !strings.HasSuffix(fs.Name(), ".mp3") {
			continue
		}

		tmpList = append(tmpList, fs.Name())
		totItems += 1
	}

	response := make([]string, totItems)
	response = tmpList

	sort.Strings(response)

	return response, nil
}

func (i *ID3Tags) GetAllRatings() (map[string]int, error) {
	ratings := map[string]int{}

	files, err := ioutil.ReadDir(i.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("ID3Tags.GetAllRatings ioutil.ReadDir: %v", err)
	}

	for _, fs := range files {
		if fs.IsDir() {
			continue
		}

		if fs.Name() == "" {
			continue
		}

		fullPath, err := i.expandFullPath(fs.Name())
		if err != nil {
			return nil, fmt.Errorf("ID3Tags.GetAllRatings: %v", err)
		}

		rating, err := i.GetRating(fullPath)
		if err != nil {
			return nil, fmt.Errorf("ID3Tags.GetAllRatings: %v", err)
		}

		ratings[fs.Name()] = rating
	}

	return ratings, nil
}

func (i *ID3Tags) GetAuthor(fname string) (string, error) {
	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.GetAuthor: %v", err)
	}

	result, err := i.ReadFrame(fullPath, ARTIST)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.GetAuthor: %v", err)
	}

	return result, nil
}

func (i *ID3Tags) GetTitle(fname string) (string, error) {
	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.GetTitle: %v", err)
	}

	result, _ := i.ReadFrame(fullPath, TITLE)
	return result, nil
}

func (i *ID3Tags) SetMetadata(fname string) error {
	var artist, title string

	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return fmt.Errorf("ID3Tags.SetMetadata: %v", err)
	}

	title, err = i.GetTitle(fullPath)
	if err != nil {
		return fmt.Errorf("ID3Tags.SetMetadata: %v", err)
	}

	if title != "" {
		return fmt.Errorf("ID3Tags.SetMetadata: Metadata already set")
	}

	rating, err := i.GetRating(fullPath)
	if err != nil {
		return fmt.Errorf("ID3Tags.SetMetadata: %v", err)
	}

	if rating != RATING_UNKNOWN {
		return fmt.Errorf("ID3Tags.SetMetadata: Rating already set already set")
	}

	name := path.Base(fullPath)
	name = name[:len(name)-16]
	rating = 5

	if strings.Count(name, "-") >= 1 {
		tokens := strings.Split(name, "-")
		artist = strings.TrimSpace(tokens[0])
		title = strings.TrimSpace(tokens[1])
	} else {
		title = name
	}

	params := []string{}
	rating_s := strconv.Itoa(rating)

	if len(artist) > 0 {
		params = []string{
			"-a", artist,
			"-t", title,
			"-T", rating_s,
			fullPath,
		}
	} else {
		params = []string{
			"-t", title,
			"-T", rating_s,
			fullPath,
		}
	}

	_, err = i.runId3v2(params)
	if err != nil {
		fmt.Errorf("ID3Tags.SetMetadata: %v", err)
	}

	return nil
}