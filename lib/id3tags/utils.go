package id3tags

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

func (i *ID3Tags) RunId3v2(params []string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("id3v2", params...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debugf("Running %s %s\n", "id3v2", strings.Join(params, " "))

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("ID3Tags.RunId3v2 cmd.Start: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		msg := fmt.Sprintf("ID3Tags.RunId3v2 cmd.Wait: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		return "", fmt.Errorf(msg)
	}

	if stderr.Len() > 0 {
		msg := fmt.Sprintf("ID3Tags.RunId3v2: failed to run: %v\n", stderr.String())
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

	output, err := i.RunId3v2(params)
	if err != nil {
		return "", fmt.Errorf("ID3Tags.ReadField: %v", err)
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

	return "", fmt.Errorf("ID3Tags.ReadField: no %s frame found", frame)
}

func (i *ID3Tags) GetId3Tags(fname string) (*Tags, error) {
	fullPath, err := i.expandFullPath(fname)
	if err != nil {
		return nil, fmt.Errorf("ID3Tags.GetId3Tags: %v", err)
	}

	params := []string{"-l", fullPath}

	output, err := i.RunId3v2(params)
	if err != nil {
		return nil, fmt.Errorf("ID3Tags.GetId3Tags: %v", err)
	}

	wantedFrames := []string{id3Artist, id3Title, id3Track, id3Comment}
	tags := &Tags{}

	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "T") {
			continue
		}

		for _, frame := range wantedFrames {
			switch frame {
			case id3Artist:
				{
					result := reArtist.FindAllStringSubmatch(line, -1)
					if len(result) == 0 {
						continue
					}
					tags.Artist = result[0][1]
				}
			case id3Title:
				{
					result := reTitle.FindAllStringSubmatch(line, -1)
					if len(result) == 0 {
						continue
					}
					tags.Title = result[0][1]
				}
			case id3Track:
				{
					result := reTrck.FindAllStringSubmatch(line, -1)
					if len(result) == 0 {
						continue
					}
					tags.Rating, err = strconv.Atoi(result[0][1])
					if err != nil {
						return nil, fmt.Errorf("ID3Tags.GetId3Tags strconv.Atoi: %v", err)
					}
				}
			case id3Comment:
				{
					result := reComm.FindAllStringSubmatch(line, -1)
					if len(result) == 0 {
						continue
					}
					tags.Comment = result[0][1]
				}
			}
		}
	}

	return tags, nil
}
