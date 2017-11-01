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

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var gTitle string = "Loading"
var gDuration string = "..."
var gRating int = -1

func (api *WebApi) updateNowPlayingData() {
	for {
		fileName := api.mpd.NowPlaying()
		if strings.HasPrefix(fileName, "Error: ") {
			fileName = api.mpd.Play()
		}
		fullPath := api.yt.MusicDir + "/" + fileName

		gTitle = fileName[:len(fileName)-16]
		gDuration = api.mpd.Duration()
		gRating = api.mp3.GetRating(fullPath)

		time.Sleep(3 * time.Second)
	}
}

func (api *WebApi) HomeHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile(api.config.Api.Assets + "/templates/player.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read template: %v\n", err)
		return
	}

	t := template.New("index")

	_, err = t.Parse(string(content))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse template: %v\n", err)
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

	err = t.Execute(w, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute template: %v\n", err)
		return
	}

	fmt.Printf("%s %s\n", r.Method, r.URL.Path)
}

func (api *WebApi) SocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to upgrade socket: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ReadMessage failed: %v\n", err)
			return
		}

		request := &ClientRequest{}
		if err := json.Unmarshal(msg, request); err != nil {
			fmt.Fprintf(os.Stderr, "Unmarshal failed: %v\n", err)
			return
		}

		switch request.Operation {
		case "np":
			{
				err = conn.WriteMessage(msgType, api.NowPlayingResponse())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
				}
			}
		case "next":
			{
				fmt.Printf("Received next\n")
				err = conn.WriteMessage(msgType, api.NextResponse())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
				}
			}
		case "boo":
			{
				fmt.Printf("Received boo\n")
				err = conn.WriteMessage(msgType, api.BooResponse())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
				}
			}
		case "tune":
			{
				fmt.Printf("Received tune\n")
				err = conn.WriteMessage(msgType, api.TuneResponse())
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
				}
			}
		case "play":
			{
				query := &SearchRequest{}
				if err := json.Unmarshal(msg, query); err != nil {
					fmt.Fprintf(os.Stderr, "Unmarshal failed: %v\n", err)
					return
				}

				pos, err := api.mpd.Search(query.Query)
				if err != nil {
					fmt.Fprintf(os.Stderr, "search failed: %v\n", err)
					return
				} else {
					fileName := api.mpd.PlayPos(pos)
					fmt.Printf("Skipping to %s", fileName[:len(fileName)-16])
				}
			}
		default:
			{
				conn.Close()
				fmt.Fprintf(os.Stderr, "Unknown message received: %s", string(msg))
				break
			}
		}
	}
}

func (api *WebApi) PlaylistHandler(w http.ResponseWriter, r *http.Request) {
	for _, file := range api.mp3.GetAllFiles() {
		if file == "" {
			continue
		}
		line := fmt.Sprintf("%s\n", file)
		w.Write([]byte(line))
	}
}

func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}
