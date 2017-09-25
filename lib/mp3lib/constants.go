package mp3lib

import "sync"

const (
	RATING_UNKNOWN   int = -1
	RATING_ZERO      int = 0
	RATING_TO_REMOVE int = 1
	RATING_MAX       int = 10

	RATING_FRAME string = "POPM"
)

type MP3Library struct {
	baseDir string
	mutex   sync.Mutex
}
