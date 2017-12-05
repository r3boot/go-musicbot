package albumart

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
)

func sha1sum(input string) (string, error) {
	hasher := sha1.New()
	nwritten, err := hasher.Write([]byte(input))
	if err != nil {
		return "", fmt.Errorf("sha1sum hasher.Write: %v", err)
	}
	if nwritten != len(input) {
		return "", fmt.Errorf("sha1sum: corrupt write")
	}

	hash := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return hash, nil
}

func SantizeQuery(q string) string {
	result := ""

	insideBrackets := false
	for _, chr := range q {
		if chr == '[' || chr == '(' || chr == '{' {
			insideBrackets = true
			continue
		}
		if chr == ']' || chr == ')' || chr == '}' {
			insideBrackets = false
			continue
		}
		if insideBrackets {
			continue
		}
		result += string(chr)
	}

	return result
}
