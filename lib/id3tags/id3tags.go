package id3tags

import (
	"fmt"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/logger"
)

var log *logger.Logger

func NewID3Tags(l *logger.Logger, cfg *config.MusicBotConfig, chanName, baseDir string, runService bool) *ID3Tags {
	log = l
	id3 := &ID3Tags{
		Config: cfg,
		BaseDir: baseDir,
		tagList: TagList{},
	}

	idxFile := fmt.Sprintf("%s/%s.bleve", cfg.Search.DataDirectory, chanName)

	err := id3.OpenSearchIndex(idxFile)
	if err != nil {
		panic(err)
	}

	if runService {
		go id3.UpdateTags()
	}

	return id3
}
