package id3tags

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
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

func (i *ID3Tags) runId3v2(params []string) (string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command("id3v2", params...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// log.Debugf("Running %s %s\n", "id3v2", strings.Join(params, " "))

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
