package utils

import (
	"fmt"
	"os"
	"regexp"
)

var (
	reYid         = regexp.MustCompile("^([a-zA-Z0-9_-]{11})$")
	reYidFilename = regexp.MustCompile(".*-([a-zA-Z0-9_-]{11}).mp3$")
)

func IsYid(value string) bool {
	results := reYid.FindAllStringSubmatch(value, -1)
	if len(results) == 0 {
		return false
	}
	return true
}

func GetYidFromFilename(fname string) (string, error) {
	results := reYidFilename.FindAllStringSubmatch(fname, -1)
	if len(results) == 0 {
		return "", fmt.Errorf("FindAllStringSubmatch: No yid found for %s", fname)
	}
	yid := results[0][1]

	return yid, nil
}

func ArgOrEnvVar(argValue interface{}, envVarName, cfgFileValue string) (interface{}, error) {
	result := argValue
	envValue, ok := os.LookupEnv(envVarName)
	if ok {
		result = envValue
	}

	if result == "" {
		result = cfgFileValue
	}

	if result == "" {
		return "", fmt.Errorf("No value found")
	}

	return result, nil
}
