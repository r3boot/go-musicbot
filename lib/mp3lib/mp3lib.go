package mp3lib

import "github.com/r3boot/go-musicbot/lib/logger"

var log *logger.Logger

func NewMP3Library(l *logger.Logger, baseDir string) *MP3Library {
	log = l

	return &MP3Library{
		BaseDir: baseDir,
	}
}
