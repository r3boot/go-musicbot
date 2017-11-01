package metrics

import (
	"fmt"
	"os"
	"strings"
)

func NumListeners() int {
	stdout, _, err := run("netstat", "-an")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NumListeners: failed to run netstat: %v", err)
		return -1
	}

	numListeners := 0
	for _, line := range stdout {
		if !strings.Contains(line, "ESTABLISHED") {
			continue
		}

		tokens := []string{}
		allTokens := strings.Split(line, " ")
		for _, token := range allTokens {
			if token == "" {
				continue
			}
			tokens = append(tokens, token)
		}

		if !strings.HasSuffix(tokens[3], ".8000") {
			continue
		}

		numListeners += 1
	}

	return numListeners
}
