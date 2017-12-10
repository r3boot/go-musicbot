package mpdclient

import (
	"strconv"
	"strings"

	"gompd/mpd"
)

func findID(entries []mpd.Attrs, title string) int {
	for _, entry := range entries {
		if !strings.Contains(entry["file"], title) {
			continue
		}

		id, err := strconv.Atoi(entry["Id"])
		if err != nil {
			log.Warningf("findID: Failed to convert int: %v", err)
			break
		}

		return id
	}
	return -1
}
