package main

import (
	"flag"
	"path"

	"fmt"

	"strings"

	"github.com/r3boot/go-musicbot/lib/id3tags"
	"github.com/r3boot/go-musicbot/lib/logger"
)

const (
	D_DEBUG     = false
	D_TIMESTAMP = false

	UNKNOWN_ARTIST = "Unknown Artist"
)

var (
	useDebug     = flag.Bool("d", D_DEBUG, "Enable debug output")
	useTimestamp = flag.Bool("t", D_TIMESTAMP, "Show timestamp in output")

	validSeparators = []string{" - ", " -- ", " _ ", "    ", "   ", "  ", " ~ ", " - ", "-", " _", "~"}

	Logger *logger.Logger
)

func FixAllId3TagsInDir(dirName string) error {
	id3 := id3tags.NewID3Tags(Logger, dirName, false)

	allFiles, err := id3.GetAllFiles()
	if err != nil {
		return fmt.Errorf("FixAllId3TagsInDir: %v", err)
	}

	for _, fname := range allFiles {
		if fname == "" {
			continue
		}

		tags, err := id3.GetId3Tags(fname)
		if err != nil {
			Logger.Warningf("FixAllId3TagsInDir: failed to read tags for %s: %v", fname, err)
		}

		if tags.Artist != "" && tags.Title != "" && tags.Rating != 0 {
			Logger.Debugf("FixAllId3TagsInDir: tags already set for %s", fname)
			continue
		}

		foundInfo := false
		artist := ""
		title := ""
		rating := 0
		name := fname[:len(fname)-16]
		fullPath := path.Join(dirName, fname)
		if err != nil {
			Logger.Warningf("FixAllId3TagsInDir filepath.Abs: %v", err)
			continue
		}

		for _, sep := range validSeparators {
			if strings.Contains(name, sep) {
				tokens := strings.Split(name, sep)
				artist = tokens[0]
				title = strings.Join(tokens[1:], sep)
				foundInfo = true
				break
			}
		}

		if !foundInfo {
			artist = UNKNOWN_ARTIST
			title = name
		}

		artist = strings.TrimSpace(artist)
		title = strings.TrimSpace(title)

		params := []string{}
		if tags.Rating == 0 {
			params = []string{
				"-a", artist,
				"-t", title,
				"-T", "5", fullPath,
				fullPath,
			}
			Logger.Debugf("params: %v", params)
		} else {
			params = []string{
				"-a", artist,
				"-t", title,
				fullPath,
			}
			rating = tags.Rating
			Logger.Debugf("params: %v", params)
		}

		Logger.Infof("Fixing id3 tags for %s (a:%s; t:%s; r:%d)", name, artist, title, rating)
		_, err = id3.RunId3v2(params)
		if err != nil {
			Logger.Warningf("FixAllId3TagsInDir: %v", err)
			continue
		}
	}

	return nil
}

func init() {
	flag.Parse()

	Logger = logger.NewLogger(*useTimestamp, *useDebug)
}

func main() {
	if flag.NArg() == 0 {
		Logger.Fatalf("Nothing to do!")
	}

	for _, dirName := range flag.Args() {
		err := FixAllId3TagsInDir(dirName)
		if err != nil {
			Logger.Fatalf("main: %v", err)
		}
	}
}
