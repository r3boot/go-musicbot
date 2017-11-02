package webapi

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"encoding/json"

	"strings"
	"time"

	"bytes"
	"github.com/gorilla/websocket"
	"sort"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var gTitle string = "Loading"
var gDuration string = "..."
var gRating int = -1

var cache CachedData

func (api *WebApi) updateNowPlayingData() {
	cache = CachedData{}

	for {
		fileName := api.mpd.NowPlaying()
		if strings.HasPrefix(fileName, "Error: ") {
			fileName = api.mpd.Play()
		}
		fullPath := api.yt.MusicDir + "/" + fileName

		cache.Title = fileName[:len(fileName)-16]
		cache.Duration = api.mpd.Duration()
		cache.Rating = api.mp3.GetRating(fullPath)

		allFiles := api.mp3.GetAllFiles()
		newList := make([]string, len(allFiles))

		for _, file := range allFiles {
			if file == "" {
				continue
			}
			newList = append(newList, file[:len(file)-16])
		}

		sort.Strings(newList)
		cache.Playlist = newList

		time.Sleep(1 * time.Second)
	}
}

func (api *WebApi) HomeHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile(api.config.Api.Assets + "/templates/player.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read template: %v\n", err)
		errmsg := "Failed to read template"
		http.Error(w, errmsg, http.StatusInternalServerError)
		httpLog(r, http.StatusInternalServerError, len(errmsg))
		return
	}

	t := template.New("index")

	_, err = t.Parse(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse template: %v\n", err)
		errmsg := "Failed to parse template"
		http.Error(w, errmsg, http.StatusInternalServerError)
		httpLog(r, http.StatusInternalServerError, len(errmsg))
		return
	}

	stream_url := api.config.Bot.StreamURL

	if api.config.Api.StreamURL != "" {
		stream_url = api.config.Api.StreamURL
	}

	data := TemplateData{
		Title:  api.config.Api.Title,
		Stream: stream_url,
	}

	output := bytes.Buffer{}

	err = t.Execute(&output, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute template: %v\n", err)
		errmsg := "Failed to execute template"
		http.Error(w, errmsg, http.StatusInternalServerError)
		httpLog(r, http.StatusInternalServerError, len(errmsg))
		return
	}

	w.Write(output.Bytes())
	httpLog(r, http.StatusOK, output.Len())
}

func (api *WebApi) SocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errmsg := fmt.Sprintf("Failed to upgrade socket: %v\n", err)
		wsLog(r, http.StatusInternalServerError, "upgrade", errmsg)
		return
	}
	defer conn.Close()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			errmsg := fmt.Sprintf("ReadMessage failed: %v\n", err)
			wsLog(r, http.StatusInternalServerError, "socket", errmsg)
			continue
		}

		request := &ClientRequest{}
		if err := json.Unmarshal(msg, request); err != nil {
			errmsg := fmt.Sprintf("Unmarshal failed: %v\n", err)
			wsLog(r, http.StatusInternalServerError, "nil", errmsg)
			continue
		}

		switch request.Operation {
		case "np":
			{
				var response = api.NowPlayingResponse()

				err = conn.WriteMessage(msgType, response)
				if err != nil {
					errmsg := fmt.Sprintf("Failed to send message: %v\n", err)
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}
				// Note, ignored
				// wsLog(r, http.StatusOK, request.Operation, "Now Playing response")
			}
		case "next":
			{
				var response = api.NowPlayingResponse()

				err = conn.WriteMessage(msgType, response)
				if err != nil {
					errmsg := fmt.Sprintf("Failed to send message: %v\n", err)
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}
				wsLog(r, http.StatusOK, request.Operation, "got nowplaying response")
			}
		case "boo":
			{

				err = conn.WriteMessage(msgType, api.BooResponse())
				if err != nil {
					errmsg := fmt.Sprintf("Failed to send message: %v\n", err)
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}
				wsLog(r, http.StatusOK, request.Operation, "got no response")
			}
		case "tune":
			{
				err = conn.WriteMessage(msgType, api.TuneResponse())
				if err != nil {
					errmsg := fmt.Sprintf("Failed to send message: %v\n", err)
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}
				wsLog(r, http.StatusOK, request.Operation, "got no response")
			}
		case "play":
			{
				query := &SearchRequest{}
				if err := json.Unmarshal(msg, query); err != nil {
					errmsg := fmt.Sprintf("Unmarshal failed: %v\n", err)
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}

				pos, err := api.mpd.Search(query.Query)
				if err != nil {
					errmsg := fmt.Sprintf("search failed: %v\n", err)
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				} else {
					fileName := api.mpd.PlayPos(pos)
					msg := fmt.Sprintf("Skipping to %s", fileName[:len(fileName)-16])
					wsLog(r, http.StatusOK, request.Operation, msg)
				}
			}
		default:
			{
				conn.Close()
				errmsg := fmt.Sprintf("Unknown message received: %s", string(msg))
				wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
				break
			}
		}
	}
}

func (api *WebApi) AutoCompleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()


	for key, value := range r.URL.Query() {
		if len(value[0]) < 3 {
			msg := "Please specify a query of 3 chars or more"
			w.Write([]byte(msg))
			httpLog(r, http.StatusOK, len(msg))
			return
		}

		q := value[0]

		results := api.mpd.TypeAheadQuery(q)

		response := AutoCompleteResponse{
			Query: q,
			Suggestions: results,
		}

		data, err := json.Marshal(response)
		if err != nil {
			errmsg := fmt.Sprintf("Failed to marshal results: %v", err)
			http.Error(w, errmsg, http.StatusInternalServerError)
			httpLog(r, http.StatusInternalServerError, len(errmsg))
			return
		}

		w.Write(data)
		httpLog(r, http.StatusOK, len(data))
		return
	}

	errmsg := "No query found"
	http.Error(w, errmsg, http.StatusInternalServerError)
	httpLog(r, http.StatusInternalServerError, len(errmsg))
}

func (api *WebApi) PlaylistHandler(w http.ResponseWriter, r *http.Request) {
	totSize := 0
	for _, title := range cache.Playlist {
		if title == "" {
			continue
		}
		line := fmt.Sprintf("%s\n", title)
		nwritten, err := w.Write([]byte(line))
		if err != nil {
			errmsg := "Failed to write playlist"
			http.Error(w, errmsg, http.StatusInternalServerError)
			httpLog(r, http.StatusInternalServerError, len(errmsg))
		}
		totSize += nwritten
	}

	httpLog(r, http.StatusOK, totSize)
}

func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpLog(r, http.StatusOK, 0)
		h.ServeHTTP(w, r)
	})
}
