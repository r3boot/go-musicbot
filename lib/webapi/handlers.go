package webapi

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"encoding/json"

	"strings"
	"time"

	"bytes"
	"sort"

	"github.com/gorilla/websocket"
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
		log.Warningf("WebApi.HomeHandler ioutil.ReadFile: %v", err)
		errmsg := "Failed to read template"
		http.Error(w, errmsg, http.StatusInternalServerError)
		httpLog(r, http.StatusInternalServerError, len(errmsg))
		return
	}

	t := template.New("index")

	_, err = t.Parse(string(content))
	if err != nil {
		log.Warningf("WebApi.HomeHandler t.Parse: %v", err)
		errmsg := "Failed to parse template"
		http.Error(w, errmsg, http.StatusInternalServerError)
		httpLog(r, http.StatusInternalServerError, len(errmsg))
		return
	}

	data := TemplateData{
		Title:     api.config.Api.Title,
		OggStream: api.config.Api.OggStreamURL,
		Mp3Stream: api.config.Api.Mp3StreamURL,
	}

	output := bytes.Buffer{}

	err = t.Execute(&output, data)
	if err != nil {
		log.Warningf("WebApi.HomeHandler t.Execute: %v", err)
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
		log.Warningf("WebApi.SocketHandler upgrader.Upgrade: %v", err)
		errmsg := fmt.Sprintf("Failed to upgrade socket: %v\n", err)
		wsLog(r, http.StatusInternalServerError, "upgrade", errmsg)
		return
	}
	defer conn.Close()

	for {
		if conn == nil {
			log.Warningf("WebApi.SocketHandler: socket closed")
			errmsg := fmt.Sprintf("Socket closed")
			wsLog(r, http.StatusInternalServerError, "socket", errmsg)
			break
		}

		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Warningf("WebApi.SocketHandler conn.ReadMessage: %v", err)
			errmsg := fmt.Sprintf("ReadMessage failed")
			wsLog(r, http.StatusInternalServerError, "socket", errmsg)
			continue
		}

		request := &ClientRequest{}
		if err := json.Unmarshal(msg, request); err != nil {
			log.Warningf("WebApi.SocketHandler json.Unmarshal: %v", err)
			errmsg := fmt.Sprintf("Unmarshal failed")
			wsLog(r, http.StatusInternalServerError, "nil", errmsg)
			continue
		}

		switch request.Operation {
		case "np":
			{
				response := api.NowPlayingResponse()

				err = conn.WriteMessage(msgType, response)
				if err != nil {
					log.Warningf("WebApi.SocketHandler conn.WriteMessage: %v", err)
					errmsg := fmt.Sprintf("Failed to send message")
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}
				// Note, ignored
				// wsLog(r, http.StatusOK, request.Operation, "Now Playing response")
			}
		case "queue":
			{
				response := api.PlayQueueResponse()

				err = conn.WriteMessage(msgType, response)
				if err != nil {
					log.Warningf("WebApi.SocketHandler conn.WriteMessage: %v", err)
					errmsg := fmt.Sprintf("Failed to send message")
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}
				// Note, ignored
				// wsLog(r, http.StatusOK, request.Operation, "Play Queue response")
			}
		case "next":
			{
				log.Debugf("WebApi.SocketHandler: got 'next' operation")

				api.mpd.Next()
				var response = api.NowPlayingResponse()

				err = conn.WriteMessage(msgType, response)
				if err != nil {
					log.Warningf("WebApi.SocketHandler conn.WriteMessage: %v", err)
					errmsg := fmt.Sprintf("Failed to send message")
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}
				wsLog(r, http.StatusOK, request.Operation, "got nowplaying response")
			}
		case "boo":
			{
				log.Debugf("WebApi.SocketHandler: got 'boo' operation")

				err = conn.WriteMessage(msgType, api.BooResponse())
				if err != nil {
					log.Warningf("WebApi.SocketHandler conn.WriteMessage: %v", err)
					errmsg := fmt.Sprintf("Failed to send message")
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}
				wsLog(r, http.StatusOK, request.Operation, "got no response")
			}
		case "tune":
			{
				log.Debugf("WebApi.SocketHandler: got 'tune' operation")

				err = conn.WriteMessage(msgType, api.TuneResponse())
				if err != nil {
					log.Warningf("WebApi.SocketHandler conn.WriteMessage: %v", err)
					errmsg := fmt.Sprintf("Failed to send message")
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}
				wsLog(r, http.StatusOK, request.Operation, "got no response")
			}
		case "request":
			{
				log.Debugf("WebApi.SocketHandler: got 'tune' operation")

				query := &SearchRequest{}
				if err := json.Unmarshal(msg, query); err != nil {
					log.Warningf("WebApi.SocketHandler json.Unmarshal: %v", err)
					errmsg := fmt.Sprintf("Unmarshal failed")
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
					continue
				}

				qpos, err := api.mpd.Enqueue(query.Query)
				if err != nil {
					log.Warningf("WebApi.SocketHandler: %v", err)
					errmsg := fmt.Sprintf("enqueue failed")
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
				}

				title, err := api.mpd.GetTitle(qpos)
				if err != nil {
					log.Warningf("WebApi.SocketHandler: %v", err)
					errmsg := fmt.Sprintf("enqueue failed")
					wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
				}

				msg := fmt.Sprintf("Added %s to the play queue", title)
				wsLog(r, http.StatusOK, request.Operation, msg)
			}
		default:
			{
				conn.Close()
				log.Warningf("WebApi.SocketHandler: unknown operation received")
				errmsg := fmt.Sprintf("Unknown operation received")
				wsLog(r, http.StatusInternalServerError, request.Operation, errmsg)
				break
			}
		}
	}
}

func (api *WebApi) AutoCompleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	for key, value := range r.URL.Query() {
		if key != "query" {
			continue
		}

		if len(value[0]) < 3 {
			msg := "Please specify a query of 3 chars or more"
			w.Write([]byte(msg))
			httpLog(r, http.StatusOK, len(msg))
			return
		}

		q := value[0]

		results := api.mpd.TypeAheadQuery(q)

		response := AutoCompleteResponse{
			Query:       q,
			Suggestions: results,
		}

		data, err := json.Marshal(response)
		if err != nil {
			log.Warningf("WebApi.AutoCompleteHandler json.Marshal: %v", err)
			errmsg := fmt.Sprintf("Failed to marshal results")
			http.Error(w, errmsg, http.StatusInternalServerError)
			httpLog(r, http.StatusInternalServerError, len(errmsg))
			return
		}

		w.Write(data)
		httpLog(r, http.StatusOK, len(data))
		return
	}

	log.Warningf("WebApi.AutoCompleteHandler: No query found")
	errmsg := "No query found"
	http.Error(w, errmsg, http.StatusInternalServerError)
	httpLog(r, http.StatusInternalServerError, len(errmsg))
}

func (api *WebApi) PlayQueueHandler(w http.ResponseWriter, r *http.Request) {
	playQueue, err := api.mpd.GetPlayQueue()
	if err != nil {
		log.Warningf("WebApi.PlayQueueHandler: %v", err)
		errmsg := fmt.Sprintf("Failed to get play queue")
		http.Error(w, errmsg, http.StatusInternalServerError)
		httpLog(r, http.StatusInternalServerError, len(errmsg))
		return
	}

	data, err := json.Marshal(playQueue)
	if err != nil {
		log.Warningf("WebApi.PlayQueueHandler json.Marshal: %v", err)
		errmsg := fmt.Sprintf("Failed to marshal json")
		http.Error(w, errmsg, http.StatusInternalServerError)
		httpLog(r, http.StatusInternalServerError, len(errmsg))
		return
	}

	w.Write(data)
	httpLog(r, http.StatusOK, len(data))
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
			log.Warningf("WebApi.PlaylistHandler w.Write: %v", err)
			errmsg := "Failed to write playlist"
			http.Error(w, errmsg, http.StatusInternalServerError)
			httpLog(r, http.StatusInternalServerError, len(errmsg))
			return
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
