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
		fullPath := api.config.Youtube.BaseDir + "/" + fileName

		gTitle = fileName[:len(fileName)-16]
		gDuration = api.mpd.Duration()
		gRating = api.mp3.GetRating(fullPath)

		fmt.Printf("np: %s (%s) %d/10\n", gTitle, gDuration, gRating)

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

	data := TemplateData{
		Title:  "2600nl radio",
		Stream: "http://radio.as65342.net:8000/2600nl.ogg",
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
		default:
			{
				conn.Close()
				fmt.Fprintf(os.Stderr, "Unknown message received: %s", string(msg))
				break
			}
		}
	}
}

func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}
