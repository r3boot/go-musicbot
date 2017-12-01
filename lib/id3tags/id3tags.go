package id3tags

import (
	"sync"

	"github.com/r3boot/go-musicbot/lib/logger"
)

var log *logger.Logger

func NewID3Tags(l *logger.Logger, baseDir string) *ID3Tags {
	log = l

	return &ID3Tags{
		BaseDir: baseDir,
		mutex:   sync.RWMutex{},
	}
}
