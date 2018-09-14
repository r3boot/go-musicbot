package id3tags

import (
	"github.com/r3boot/go-musicbot/lib/config"
	"math"
	"regexp"
	"sync"

	"github.com/blevesearch/bleve"
)

const (
	RATING_UNKNOWN   = -1
	RATING_ZERO      = 0
	RATING_TO_REMOVE = 1
	RATING_DEFAULT   = 5
	RATING_MAX       = math.MaxInt32

	id3Track   = "TRCK"
	id3Artist  = "TPE1"
	id3Title   = "TIT2"
	id3Comment = "COMM"
)

type Tags struct {
	Artist  string
	Title   string
	Rating  int
	Comment string
}

type TrackTags struct {
	Artist string
	Title  string
	YID string
	Filename string
}

type TagList map[string]*TrackTags

type ID3Tags struct {
	Config   *config.MusicBotConfig
	BaseDir   string
	tagList   TagList
	searchIdx bleve.Index
}

var (
	// id3v1 tag info for /music/2600nl/Zero 7 - In The Waiting Line-5tZlu4wP4pw.mp3:
	reTrack = regexp.MustCompile("^id3v1 tag info for (.*):$")

	// TPE1 (Lead performer(s)/Soloist(s)): Zero 7
	reArtist = regexp.MustCompile("^TPE1 .*: (.*)$")

	// TIT2 (Title/songname/content description): In The Waiting Line
	reTitle = regexp.MustCompile("^TIT2 .*: (.*)$")

	reTrck = regexp.MustCompile("^TRCK .*: (.*)$")

	reComm = regexp.MustCompile("^COMM .*\\[\\]: ([a-zA-Z0-9_\\-]+)$")

	tagListMutex sync.RWMutex
)
