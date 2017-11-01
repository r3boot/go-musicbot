package metrics

import (
	"bytes"
	"os/exec"
	"strings"
)

func run(command string, args ...string) (stdout, stderr []string, err error) {
	var stdout_buf, stderr_buf bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout_buf
	cmd.Stderr = &stderr_buf

	if err = cmd.Run(); err != nil {
		return
	}

	stdout = strings.Split(stdout_buf.String(), "\n")
	stderr = strings.Split(stderr_buf.String(), "\n")

	return
}
