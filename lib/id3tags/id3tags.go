package id3tags

import "github.com/r3boot/go-musicbot/lib/logger"

var log *logger.Logger

func NewID3Tags(l *logger.Logger, baseDir string) *ID3Tags {
	log = l
	id3 := &ID3Tags{
		BaseDir: baseDir,
		tagList: TagList{},
	}

	go id3.UpdateTags()

	return id3
}
