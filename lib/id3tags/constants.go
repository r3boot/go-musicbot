package id3tags

import (
	"regexp"
	"sync"
)

const (
	RATING_UNKNOWN   int = -1
	RATING_ZERO      int = 0
	RATING_TO_REMOVE int = 1
	RATING_DEFAULT   int = 5
	RATING_MAX       int = 10

	TRACK  string = "TRCK"
	ARTIST string = "TPE1"
	TITLE  string = "TIT2"
)

type ID3Tags struct {
	BaseDir string
	mutex   sync.RWMutex
}

var (
	RE_TRACK  = regexp.MustCompile("^TRCK.*: ([0-9]+)$")
	RE_ARTIST = regexp.MustCompile("^TPE1.*: (.*)$")
	RE_TITLE  = regexp.MustCompile("^TIT2.*: (.*)$")
)
