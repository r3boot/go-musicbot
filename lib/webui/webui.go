package webui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/r3boot/test/models"

	"github.com/gorilla/websocket"

	"github.com/go-openapi/runtime"
	"github.com/sirupsen/logrus"

	"github.com/r3boot/test/lib/apiclient"
	"github.com/r3boot/test/lib/apiclient/operations"
	"github.com/r3boot/test/lib/config"
)

const (
	timeFormatClf = "02/Jan/2006:15:04:05 -0700"

	wsRequestNowPlaying = 0
	wsRequestNext       = 1
	wsRequestBoo        = 2
	wsRequestTune       = 3
	wsRequestSearch     = 4
	wsRequestTrack      = 5
	wsRequestQueue      = 6

	wsReplyNowPlaying = 100
	wsReplyNext       = 101
	wsReplyBoo        = 102
	wsReplyTune       = 103
	wsReplySearch     = 104
	wsReplyTrack      = 105
	wsReplyQueue      = 106

	wsTextMessage = 1

	maxQueueSize = 10

	maxSearchResults = 1

	perClientSendBufferSize = 256
	maxMessageSize          = 4096
	pingTimeout             = 1
	pongTimeout             = 1
	writeTimeout            = 1
)

type WebUi struct {
	uri          string
	token        runtime.ClientAuthInfoWriter
	client       *apiclient.Musicbot
	nowPlaying   *models.Track
	queueEntries map[int]string
}

type AccessLogEntry struct {
	SrcIp     string
	Timestamp string
	Method    string
	Path      string
	Proto     string
	Code      int
	size      int
}

type wsClientRequest struct {
	Operation int         `json:"o"`
	Data      interface{} `json:"d"`
}

type wsServerResponse struct {
	Operation int         `json:"o"`
	Data      interface{} `json:"d"`
	Status    bool        `json:"s"`
	Message   string      `json:"m"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func NewWebUi(config *config.WebUi, token runtime.ClientAuthInfoWriter, client *apiclient.Musicbot) (*WebUi, error) {
	uri := fmt.Sprintf("%s:%d", config.Address, config.Port)
	ui := &WebUi{
		uri:    uri,
		token:  token,
		client: client,
	}

	go ui.fetchNowPlaying()
	go ui.fetchPlayQueue()

	return ui, nil
}

func (ui *WebUi) fetchNowPlaying() {
	log := logrus.WithFields(logrus.Fields{
		"module":   "WebUi",
		"function": "fetchNowPlaying",
	})
	for {
		resp, err := ui.client.Operations.GetPlayerNowplaying(operations.NewGetPlayerNowplayingParams(), ui.token)
		if err != nil {
			log.Warningf("Failed to fetch nowplaying info: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}
		track := models.Track{
			Addedon:   resp.Payload.Addedon,
			Duration:  resp.Payload.Duration,
			Elapsed:   resp.Payload.Elapsed,
			Filename:  resp.Payload.Filename,
			ID:        resp.Payload.ID,
			Priority:  resp.Payload.Priority,
			Rating:    resp.Payload.Rating,
			Submitter: resp.Payload.Submitter,
		}

		ui.nowPlaying = &track
		time.Sleep(2000 * time.Millisecond)
	}
}

func (ui *WebUi) fetchPlayQueue() {
	log := logrus.WithFields(logrus.Fields{
		"module":   "WebUi",
		"function": "fetchPlayQueue",
	})
	for {
		response, err := ui.client.Operations.GetPlayerQueue(operations.NewGetPlayerQueueParams(), ui.token)
		if err != nil {
			log.Warningf("Failed to fetch play queue: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		//queueEntries := response.Payload
		queueEntries := make(map[int]string)

		for idx, entry := range response.Payload {
			queueEntries[idx] = *entry.Filename
		}

		ui.queueEntries = queueEntries
		time.Sleep(2000 * time.Millisecond)
	}
}

func (ui *WebUi) NewResponse(responseType int, data interface{}, status bool, msg string) wsServerResponse {
	response := wsServerResponse{
		Operation: responseType,
		Data:      data,
		Status:    status,
		Message:   msg,
	}

	return response
}

func (ui *WebUi) NewSuccesResponse(msgType int, content interface{}) ([]byte, error) {
	response := ui.NewResponse(msgType, content, true, "")
	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %v", err)
	}

	return data, nil
}

func (ui *WebUi) NewFailedResponse(msgType int, msg string) ([]byte, error) {
	response := ui.NewResponse(msgType, nil, false, "")
	data, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %v", err)
	}

	return data, nil
}

func (ui *WebUi) wsSend(conn *websocket.Conn, data []byte) error {
	err := conn.WriteMessage(wsTextMessage, data)
	if err != nil {
		return fmt.Errorf("conn.WriteMessage: %v", err)
	}
	return nil
}

func (ui *WebUi) HandleNext(conn *websocket.Conn, r *http.Request) error {
	resp, err := ui.client.Operations.GetPlayerNext(operations.NewGetPlayerNextParams(), ui.token)
	if err != nil {
		response, err := ui.NewFailedResponse(wsReplyNext, "Failed to skip to next track")
		if err != nil {
			return fmt.Errorf("ui.NewFailedResponse: %v", err)
		}

		err = ui.wsSend(conn, response)
		if err != nil {
			return fmt.Errorf("ui.wsSend: %v", err)
		}
	}

	track := resp.GetPayload()

	response, err := ui.NewSuccesResponse(wsReplyNext, track)
	if err != nil {
		return fmt.Errorf("ui.NewSuccesResponse: %v", err)

	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("ui.wsSend: %v", err)
	}

	return nil
}

func (ui *WebUi) HandleBoo(conn *websocket.Conn, r *http.Request) error {
	resp, err := ui.client.Operations.GetRatingDecrease(operations.NewGetRatingDecreaseParams(), ui.token)
	if err != nil {
		response, err := ui.NewFailedResponse(wsReplyBoo, "Failed to boo track")
		if err != nil {
			return fmt.Errorf("ui.NewFailedResponse: %v", err)
		}

		err = ui.wsSend(conn, response)
		if err != nil {
			return fmt.Errorf("ui.wsSend: %v", err)
		}
	}

	track := resp.GetPayload()

	response, err := ui.NewSuccesResponse(wsReplyBoo, track)
	if err != nil {
		return fmt.Errorf("ui.NewSuccesResponse: %v", err)

	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("ui.wsSend: %v", err)
	}

	return nil
}

func (ui *WebUi) HandleTune(conn *websocket.Conn, r *http.Request) error {
	resp, err := ui.client.Operations.GetRatingIncrease(operations.NewGetRatingIncreaseParams(), ui.token)
	if err != nil {
		response, err := ui.NewFailedResponse(wsReplyTune, "Failed to boo track")
		if err != nil {
			return fmt.Errorf("ui.NewFailedResponse: %v", err)
		}

		err = ui.wsSend(conn, response)
		if err != nil {
			return fmt.Errorf("ui.wsSend: %v", err)
		}
	}

	track := resp.GetPayload()

	response, err := ui.NewSuccesResponse(wsReplyTune, track)
	if err != nil {
		return fmt.Errorf("ui.NewSuccesResponse: %v", err)

	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("ui.wsSend: %v", err)
	}

	return nil
}

func (ui *WebUi) HandleNowPlaying(conn *websocket.Conn, r *http.Request) error {
	if conn == nil {
		return fmt.Errorf("Connection closed")
	}

	response, err := ui.NewSuccesResponse(wsReplyNowPlaying, ui.nowPlaying)
	if err != nil {
		return fmt.Errorf("ui.NewSuccessResponse: %v", err)
	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("ui.wsSend: %v", err)
	}
	return nil
}

func (ui *WebUi) HandleSearch(conn *websocket.Conn, r *http.Request, query string) error {
	if conn == nil {
		return fmt.Errorf("Connection closed")
	}

	source := ui.getSource(r)

	params := operations.NewPostTrackSearchParams()
	params.Request = operations.PostTrackSearchBody{
		Query:     &query,
		Submitter: &source,
	}

	response, err := ui.client.Operations.PostTrackSearch(params, ui.token)
	if err != nil {
		return fmt.Errorf("PostTrackSearch: %v", err)
	}

	result := []string{}
	for _, track := range response.Payload {
		result = append(result, string(*track.Filename))
	}

	reply, err := ui.NewSuccesResponse(wsReplySearch, result)
	if err != nil {
		return fmt.Errorf("ui.NewSuccessResponse: %v", err)
	}

	err = ui.wsSend(conn, reply)
	if err != nil {
		return fmt.Errorf("ui.wsSend: %v", err)
	}
	return nil
}

func (ui *WebUi) HandleRequestTrack(conn *websocket.Conn, r *http.Request, query string) error {
	if conn == nil {
		return fmt.Errorf("Connection closed")
	}

	source := ui.getSource(r)

	params := operations.NewPostTrackRequestParams()
	params.Request = operations.PostTrackRequestBody{
		Query:     &query,
		Submitter: &source,
	}

	response, err := ui.client.Operations.PostTrackRequest(params, ui.token)
	if err != nil {
		return fmt.Errorf("PostTrackSearch: %v", err)
	}

	reply, err := ui.NewSuccesResponse(wsReplyTrack, response.Payload.Track.Priority)
	if err != nil {
		return fmt.Errorf("ui.NewSuccessResponse: %v", err)
	}

	err = ui.wsSend(conn, reply)
	if err != nil {
		return fmt.Errorf("ui.wsSend: %v", err)
	}
	return nil
}

func (ui *WebUi) HandleQueue(conn *websocket.Conn, r *http.Request) error {
	if conn == nil {
		return fmt.Errorf("Connection closed")
	}

	response, err := ui.NewSuccesResponse(wsReplyQueue, ui.queueEntries)
	if err != nil {
		return fmt.Errorf("ui.NewSuccessResponse: %v", err)
	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("ui.wsSend: %v", err)
	}
	return nil
}

func (ui *WebUi) getSource(r *http.Request) string {
	source := r.Header.Get("X-Forwarded-For")
	if source == "" {
		source = r.RemoteAddr
	}

	return source
}

func (ui *WebUi) accessLogEntry(r *http.Request, code, size int) {
	source := ui.getSource(r)

	logline := source + " - - [" + time.Now().Format(timeFormatClf) + "] "
	logline += "\"" + r.Method + " " + r.URL.Path + " " + r.Proto + "\" "
	logline += strconv.Itoa(code) + " " + strconv.Itoa(size)

	fmt.Printf("%s\n", logline)
}

func (ui *WebUi) wsAccessLogEntry(r *http.Request, code int, cmd string) {
	source := ui.getSource(r)

	logline := source + " [" + time.Now().Format(timeFormatClf) + "] "
	logline += strconv.Itoa(code) + " " + cmd

	fmt.Printf("%s\n", logline)
}

func (ui *WebUi) Run() error {
	log := logrus.WithFields(logrus.Fields{
		"module":   "WebUi",
		"function": "Run",
	})

	http.HandleFunc("/", ui.AssetHandler)
	http.HandleFunc("/ws", ui.SocketHandler)

	log.Infof("Listening on %s", ui.uri)
	err := http.ListenAndServe(ui.uri, nil)
	if err != nil {
		return fmt.Errorf("WebApi.Run http.ListenAndServe: %v", err)
	}
	return nil
}

func (ui *WebUi) AssetHandler(w http.ResponseWriter, r *http.Request) {
	log := logrus.WithFields(logrus.Fields{
		"module":   "WebUi",
		"function": "AssetHandler",
	})
	switch r.Method {
	case http.MethodGet:
		{
			path := ""
			if r.URL.Path == "/" {
				path = "html_index_html"
			} else {
				path = strings.ReplaceAll(r.URL.Path[1:], "/", "_")
				path = strings.ReplaceAll(path, ".", "_")
			}

			content, ok := webuiAssets[path]
			if !ok {
				log.Warningf("No assets found for %s, expected %s", r.URL.Path, path)
				http.NotFound(w, r)
				ui.accessLogEntry(r, 0, http.StatusNotFound)
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(content.Data)
			if err != nil {
				log.Warningf("Failed to decode base64 for %s: %v", path, err)
				http.Error(w, "", http.StatusInternalServerError)
				ui.accessLogEntry(r, 0, http.StatusInternalServerError)
				return
			}

			w.Header().Set("ContentType", content.ContentType)

			output := bytes.Buffer{}
			output.Write(decoded)

			w.Write(output.Bytes())
			ui.accessLogEntry(r, len(decoded), http.StatusOK)
			return
		}
	default:
		{
			log.Warningf("Unsupported method")
			http.Error(w, "", 405)
			ui.accessLogEntry(r, 0, 405)
			return
		}
	}

	log.Warningf("File not found")
	http.NotFound(w, r)
	ui.accessLogEntry(r, 0, 404)
}

func (ui *WebUi) SocketHandler(w http.ResponseWriter, r *http.Request) {
	log := logrus.WithFields(logrus.Fields{
		"module":   "WebUi",
		"function": "AssetHandler",
	})

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warningf("upgrader.Upgrade: %v", err)
		http.Error(w, "", 500)
		ui.accessLogEntry(r, 0, 500)
		return
	}

	source := ui.getSource(r)

	log = logrus.WithFields(logrus.Fields{
		"module":   "WebUi",
		"function": "AssetHandler",
		"source":   source,
	})

	for {
		if conn == nil {
			ui.wsAccessLogEntry(r, 500, "socket closed")
			break
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Warningf("conn.ReadMessage: %v", err)
			conn = nil
			break
		}

		request := &wsClientRequest{}
		if err := json.Unmarshal(msg, request); err != nil {
			log.Warningf("json.Unmarshal: %v", err)
			log.Debugf("%v", string(msg))
			ui.wsAccessLogEntry(r, 500, "socket closed")
			continue
		}

		switch request.Operation {
		case wsRequestNowPlaying:
			{
				err = ui.HandleNowPlaying(conn, r)
				if err != nil {
					msg := fmt.Sprintf("HandleNowPlaying: %v", err)
					response, err := ui.NewFailedResponse(wsReplyNowPlaying, msg)
					if err != nil {
						log.Warningf("ui.NewFailedResponse: %v", err)
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf("ui.wsSend: %v", err)
						break
					}
				}
			}
		case wsRequestNext:
			{
				err = ui.HandleNext(conn, r)
				if err != nil {
					msg := fmt.Sprintf("HandleNext: %v", err)
					response, err := ui.NewFailedResponse(wsReplyNext, msg)
					if err != nil {
						log.Warningf("ui.NewFailedResponse: %v", err)
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf("ui.wsSend: %v", err)
						break
					}
				}
			}
		case wsRequestBoo:
			{
				err = ui.HandleBoo(conn, r)
				if err != nil {
					msg := fmt.Sprintf("HandleBoo: %v", err)
					response, err := ui.NewFailedResponse(wsReplyBoo, msg)
					if err != nil {
						log.Warningf("ui.NewFailedResponse: %v", err)
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf("ui.wsSend: %v", err)
						break
					}
				}
			}
		case wsRequestTune:
			{
				err = ui.HandleTune(conn, r)
				if err != nil {
					msg := fmt.Sprintf("HandleTune: %v", err)
					response, err := ui.NewFailedResponse(wsReplyTune, msg)
					if err != nil {
						log.Warningf("ui.NewFailedResponse: %v", err)
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf("ui.wsSend: %v", err)
						break
					}
				}
			}
		case wsRequestSearch:
			{
				err = ui.HandleSearch(conn, r, request.Data.(string))
				if err != nil {
					msg := fmt.Sprintf("HandleSearch: %v", err)
					response, err := ui.NewFailedResponse(wsReplySearch, msg)
					if err != nil {
						log.Warningf("ui.NewFailedResponse: %v", err)
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf("ui.wsSend: %v", err)
						break
					}
				}
			}
		case wsRequestTrack:
			{
				err = ui.HandleRequestTrack(conn, r, request.Data.(string))
				if err != nil {
					msg := fmt.Sprintf("HandleRequestTrack: %v", err)
					response, err := ui.NewFailedResponse(wsReplyTrack, msg)
					if err != nil {
						log.Warningf("ui.NewFailedResponse: %v", err)
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf("ui.wsSend: %v", err)
						break
					}
				}
			}
		case wsRequestQueue:
			{
				err = ui.HandleQueue(conn, r)
				if err != nil {
					msg := fmt.Sprintf("HandleQueue: %v", err)
					response, err := ui.NewFailedResponse(wsReplyQueue, msg)
					if err != nil {
						log.Warningf("ui.NewFailedResponse: %v", err)
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf("ui.wsSend: %v", err)
						break
					}
				}
			}
		default:
			{
				conn.Close()
				log.Warningf("WebApi.SocketHandler: unknown operation received: %d", request.Operation)
				ui.wsAccessLogEntry(r, 500, fmt.Sprintf("%d", request.Operation))
				break
			}
		}
	}
}
