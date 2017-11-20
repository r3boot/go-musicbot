package mp3lib

import "sync"

const MAX_PLAYLIST_LENGTH int = 8192

const (
	RATING_UNKNOWN   int = -1
	RATING_ZERO      int = 0
	RATING_TO_REMOVE int = 1
	RATING_DEFAULT   int = 5
	RATING_MAX       int = 10

	RATING_FRAME string = "POPM"
)

type RatingsMap map[string]int

type MP3Library struct {
	BaseDir string
	mutex   sync.Mutex
}
