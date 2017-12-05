package albumart

import (
	"fmt"

	"github.com/r3boot/go-musicbot/lib/logger"
)

var log *logger.Logger

func NewAlbumArt(l *logger.Logger, webAssets string) *AlbumArt {
	log = l

	cacheDir := fmt.Sprintf("%s/img/art", webAssets)

	return &AlbumArt{
		cacheDir: cacheDir,
	}
}
