package metrics

import (
	"fmt"
	"os"
	"strings"
)

func NumListeners() int {
	stdout, _, err := run("ss", "-nte")
	if err != nil {
		fmt.Fprintf(os.Stderr, "NumListeners: failed to run ss: %v", err)
		return -1
	}

	numListeners := 0
	for _, line := range stdout {
		if strings.HasPrefix(line, "State") {
			continue
		}

		tokens := []string{}
		allTokens := strings.Split(line, " ")
		for _, t := range allTokens {
			if len(t) == 0 {
				continue
			}
			tokens = append(tokens, t)
		}

		if len(tokens) == 0 {
			continue
		}

		if strings.HasSuffix(tokens[3], ":8666") {

			numListeners += 1
		}
	}

	return numListeners
}
