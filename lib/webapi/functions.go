package webapi

import (
	"net/http"
	"time"

	"encoding/json"
	"fmt"

	"sort"

	"github.com/r3boot/go-musicbot/lib/mpdclient"
)

func (r WebResponse) ToJSON() ([]byte, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("WebResponse.ToJSON json.Marshal: %v", err)
	}

	return data, nil
}

func (a *WebAPI) UpdatePlaylist() {
	for {
		tStart := time.Now()

		playlistEntries, err := a.mpdClient.GetPlaylist()
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Debugf("WebAPI.UpdatePlaylist: Got %d playlist entries", len(playlistEntries))

		id3Entries, err := a.id3Tags.GetTags()
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Debugf("WebAPI.UpdatePlaylist: Got %d entries with id3 tags", len(id3Entries))

		playlist := mpdclient.Playlist{}
		artists := mpdclient.Artists{}

		for _, entry := range playlistEntries {
			if entry == nil {
				continue
			}

			if !id3Entries.Has(entry.Filename) {
				playlist[entry.Filename] = entry
				continue
			}

			entry.Artist = id3Entries[entry.Filename].Artist
			entry.Title = id3Entries[entry.Filename].Title
			playlist[entry.Filename] = entry

			if entry.Artist != "" && !artists.Has(entry.Artist) {
				artists = append(artists, entry.Artist)
			} else if entry.Title != "" {
				artists = append(artists, entry.Title)
			}
		}

		sort.Sort(artists)

		a.Playlist = playlist
		a.Artists = artists

		tDuration := time.Since(tStart)
		log.Debugf("Updated playlist with %d entries in %v", len(playlist), tDuration)
		time.Sleep((5 * time.Second) - tDuration)
	}
}

func (a *WebAPI) Run() error {
	// Web API
	http.HandleFunc("/api/v1/playlist", a.PlaylistHandler)
	http.HandleFunc("/api/v1/artist", a.ArtistHandler)
	http.HandleFunc("/ws", a.SocketHandler)

	// Website
	http.Handle("/css/", logHandler(http.FileServer(http.Dir(a.assets))))
	http.Handle("/js/", logHandler(http.FileServer(http.Dir(a.assets))))
	http.Handle("/img/", logHandler(http.FileServer(http.Dir(a.assets))))
	http.Handle("/fonts/", logHandler(http.FileServer(http.Dir(a.assets))))
	http.HandleFunc("/", a.HomeHandler)

	log.Infof("Listening on %s", a.address)
	err := http.ListenAndServe(a.address, nil)
	if err != nil {
		return fmt.Errorf("WebApi.Run http.ListenAndServe: %v", err)
	}

	return nil
}
