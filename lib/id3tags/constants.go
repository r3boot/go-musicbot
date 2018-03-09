package id3tags

import (
	"math"
	"regexp"
	"sync"
)

const (
	RATING_UNKNOWN   int = -1
	RATING_ZERO      int = 0
	RATING_TO_REMOVE int = 1
	RATING_DEFAULT   int = 5
	RATING_MAX       int = math.MaxInt32

	TRACK  string = "TRCK"
	ARTIST string = "TPE1"
	TITLE  string = "TIT2"
)

type Tags struct {
	Artist string
	Title  string
	Rating int
}

type TrackTags struct {
	Artist string
	Title  string
}

type TagList map[string]*TrackTags

type ID3Tags struct {
	BaseDir string
	tagList TagList
}

var (
	// id3v1 tag info for /music/2600nl/Zero 7 - In The Waiting Line-5tZlu4wP4pw.mp3:
	RE_TRACK = regexp.MustCompile("^id3v1 tag info for (.*):$")

	// TPE1 (Lead performer(s)/Soloist(s)): Zero 7
	RE_ARTIST = regexp.MustCompile("^TPE1 .*: (.*)$")

	// TIT2 (Title/songname/content description): In The Waiting Line
	RE_TITLE = regexp.MustCompile("^TIT2 .*: (.*)$")

	RE_TRCK = regexp.MustCompile("^TRCK .*: (.*)$")

	tagListMutex sync.RWMutex
)
