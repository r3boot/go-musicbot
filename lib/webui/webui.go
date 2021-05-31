package webui

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/r3boot/go-musicbot/lib/dbclient"
	"github.com/r3boot/go-musicbot/lib/discogs"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/go-openapi/runtime"

	"github.com/r3boot/go-musicbot/lib/apiclient"
	"github.com/r3boot/go-musicbot/lib/apiclient/operations"
	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/log"
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
	cfg          *config.WebUi
	nowPlaying   *dbclient.Track
	albumArt     string
	queueEntries map[int]string
	discogs      *discogs.Discogs
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

func NewWebUi(cfg *config.WebUi, token runtime.ClientAuthInfoWriter, client *apiclient.Musicbot) (*WebUi, error) {
	uri := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)
	ui := &WebUi{
		uri:    uri,
		token:  token,
		client: client,
		cfg:    cfg,
	}

	ui.discogs = discogs.NewDiscogs(ui.cfg.Discogs)

	go ui.fetchNowPlaying()
	go ui.fetchPlayQueue()

	return ui, nil
}

func (ui *WebUi) fetchNowPlaying() {
	currentFilename := ""
	for {
		resp, err := ui.client.Operations.GetPlayerNowplaying(operations.NewGetPlayerNowplayingParams(), ui.token)
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "webui",
				"function": "fetchNowPlaying",
				"call":     "ui.client.Operations.GetPlayerNowplaying",
			}, err.Error())
			time.Sleep(3 * time.Second)
			continue
		}

		addedon, err := time.Parse("2006-01-02 15:04:05 +0000 MST", *resp.Payload.Addedon)
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "webui",
				"function":  "fetchNowPlaying",
				"call":      "time.Parse",
				"timestamp": *resp.Payload.Addedon,
			}, err.Error())
			continue
		}
		duration := float64(*resp.Payload.Duration)
		elapsed := time.Duration(*resp.Payload.Elapsed)

		track := dbclient.Track{
			AddedOn:   addedon,
			Duration:  duration,
			Elapsed:   elapsed,
			Filename:  *resp.Payload.Filename,
			Rating:    *resp.Payload.Rating,
			Submitter: *resp.Payload.Submitter,
		}

		if currentFilename != track.Filename {
			query := track.Filename[:len(track.Filename)-16]
			imgFilename, err := ui.discogs.GetAlbumArt(query)
			if err != nil {
				log.Warningf(log.Fields{
					"package":  "webui",
					"function": "fetchNowPlaying",
					"call":     "ui.discogs.GetAlbumArt",
				}, err.Error())
			}
			ui.albumArt = imgFilename

			currentFilename = track.Filename
		}

		ui.nowPlaying = &track

		time.Sleep(2000 * time.Millisecond)
	}
}

func (ui *WebUi) fetchPlayQueue() {
	for {
		response, err := ui.client.Operations.GetPlayerQueue(operations.NewGetPlayerQueueParams(), ui.token)
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "webui",
				"function": "fetchPlayQueue",
				"call":     "GetPlayerQueue",
			}, err.Error())
			time.Sleep(3 * time.Second)
			continue
		}

		queueEntries := make(map[int]string)

		for idx, entry := range response.Payload {
			queueEntries[idx] = *entry.Filename
		}

		ui.queueEntries = queueEntries
		time.Sleep(1000 * time.Millisecond)
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
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "NewSuccesResponse",
			"call":     "json.Marshal",
		}, err.Error())
		return nil, fmt.Errorf("failed to marshal json")
	}

	return data, nil
}

func (ui *WebUi) NewFailedResponse(msgType int, msg string) ([]byte, error) {
	response := ui.NewResponse(msgType, nil, false, "")
	data, err := json.Marshal(response)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "NewFailedResponse",
			"call":     "json.Marshal",
		}, err.Error())
		return nil, fmt.Errorf("json.Marshal: %v", err)
	}

	return data, nil
}

func (ui *WebUi) wsSend(conn *websocket.Conn, data []byte) error {
	err := conn.WriteMessage(wsTextMessage, data)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "wsSend",
			"call":     "conn.WriteMessage",
		}, err.Error())
		return fmt.Errorf("failed to send on websocket")
	}
	return nil
}

func (ui *WebUi) HandleNext(conn *websocket.Conn, r *http.Request) error {
	resp, err := ui.client.Operations.GetPlayerNext(operations.NewGetPlayerNextParams(), ui.token)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "HandleNext",
			"call":     "ui.client.Operations.GetPlayerNext",
		}, err.Error())

		response, err := ui.NewFailedResponse(wsReplyNext, "Failed to skip to next track")
		if err != nil {
			return fmt.Errorf("NewFailedResponse: %v", err)
		}

		err = ui.wsSend(conn, response)
		if err != nil {
			return fmt.Errorf("wsSend: %v", err)
		}
	}

	track := resp.GetPayload()

	response, err := ui.NewSuccesResponse(wsReplyNext, track)
	if err != nil {
		return fmt.Errorf("NewSuccesResponse: %v", err)

	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("wsSend: %v", err)
	}

	return nil
}

func (ui *WebUi) HandleBoo(conn *websocket.Conn, r *http.Request) error {
	resp, err := ui.client.Operations.GetRatingDecrease(operations.NewGetRatingDecreaseParams(), ui.token)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "HandleBoo",
			"call":     "ui.client.Operations.GetRatingDecrease",
		}, err.Error())

		response, err := ui.NewFailedResponse(wsReplyBoo, "failed to downvote track")
		if err != nil {
			return fmt.Errorf("NewFailedResponse: %v", err)
		}

		err = ui.wsSend(conn, response)
		if err != nil {
			return fmt.Errorf("wsSend: %v", err)
		}
	}

	track := resp.GetPayload()

	response, err := ui.NewSuccesResponse(wsReplyBoo, track)
	if err != nil {
		return fmt.Errorf("NewSuccesResponse: %v", err)
	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("wsSend: %v", err)
	}

	return nil
}

func (ui *WebUi) HandleTune(conn *websocket.Conn, r *http.Request) error {
	resp, err := ui.client.Operations.GetRatingIncrease(operations.NewGetRatingIncreaseParams(), ui.token)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "HandleTune",
			"call":     "ui.client.Operations.GetRatingIncrease",
		}, err.Error())

		response, err := ui.NewFailedResponse(wsReplyTune, "Failed to upvote track")
		if err != nil {
			return fmt.Errorf("NewFailedResponse: %v", err)
		}

		err = ui.wsSend(conn, response)
		if err != nil {
			return fmt.Errorf("wsSend: %v", err)
		}
	}

	track := resp.GetPayload()

	response, err := ui.NewSuccesResponse(wsReplyTune, track)
	if err != nil {
		return fmt.Errorf("NewSuccesResponse: %v", err)

	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("wsSend: %v", err)
	}

	return nil
}

func (ui *WebUi) HandleNowPlaying(conn *websocket.Conn, r *http.Request) error {
	if conn == nil {
		return fmt.Errorf("connection closed")
	}

	ui.nowPlaying.AlbumArt = ui.albumArt

	response, err := ui.NewSuccesResponse(wsReplyNowPlaying, ui.nowPlaying)
	if err != nil {
		return fmt.Errorf("NewSuccessResponse: %v", err)
	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("wsSend: %v", err)
	}
	return nil
}

func (ui *WebUi) HandleSearch(conn *websocket.Conn, r *http.Request, query string) error {
	if conn == nil {
		return fmt.Errorf("connection closed")
	}

	source := ui.getSource(r)

	params := operations.NewPostTrackSearchParams()
	params.Request = operations.PostTrackSearchBody{
		Query:     &query,
		Submitter: &source,
	}

	response, err := ui.client.Operations.PostTrackSearch(params, ui.token)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "HandleSearch",
			"call":     "ui.client.Operations.PostTrackSearch",
		}, err.Error())
		return fmt.Errorf("failed to search")
	}

	result := []string{}
	for _, track := range response.Payload {
		result = append(result, *track.Filename)
	}

	reply, err := ui.NewSuccesResponse(wsReplySearch, result)
	if err != nil {
		return fmt.Errorf("NewSuccessResponse: %v", err)
	}

	err = ui.wsSend(conn, reply)
	if err != nil {
		return fmt.Errorf("wsSend: %v", err)
	}
	return nil
}

func (ui *WebUi) HandleRequestTrack(conn *websocket.Conn, r *http.Request, query string) error {
	if conn == nil {
		return fmt.Errorf("connection closed")
	}

	source := ui.getSource(r)

	params := operations.NewPostTrackRequestParams()
	params.Request = operations.PostTrackRequestBody{
		Query:     &query,
		Submitter: &source,
	}

	response, err := ui.client.Operations.PostTrackRequest(params, ui.token)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "HandleRequestTrack",
			"call":     "ui.client.Operations.PostTrackRequest",
		}, err.Error())
		return fmt.Errorf("failed to request track")
	}

	reply, err := ui.NewSuccesResponse(wsReplyTrack, response.Payload.Track.Priority)
	if err != nil {
		return fmt.Errorf("NewSuccessResponse: %v", err)
	}

	err = ui.wsSend(conn, reply)
	if err != nil {
		return fmt.Errorf("wsSend: %v", err)
	}
	return nil
}

func (ui *WebUi) HandleQueue(conn *websocket.Conn, r *http.Request) error {
	if conn == nil {
		return fmt.Errorf("connection closed")
	}

	response, err := ui.NewSuccesResponse(wsReplyQueue, ui.queueEntries)
	if err != nil {
		return fmt.Errorf("NewSuccessResponse: %v", err)
	}

	err = ui.wsSend(conn, response)
	if err != nil {
		return fmt.Errorf("wsSend: %v", err)
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", ui.AssetHandler)
	mux.HandleFunc("/ws", ui.SocketHandler)

	log.Debugf(log.Fields{
		"package":  "webui",
		"function": "Run",
		"uri":      ui.uri,
	}, "serving webui")

	err := http.ListenAndServe(ui.uri, mux)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "webui",
			"function": "Run",
			"call":     "http.ListenAndServe",
		}, err.Error())
		return fmt.Errorf("failed to run api")
	}
	return nil
}

func (ui *WebUi) fetchGeneratedAsset(uri string) ([]byte, string, error) {
	path := ""
	if uri == "/" {
		path = "html_index_html"
	} else {
		path = strings.ReplaceAll(uri[1:], "/", "_")
		path = strings.ReplaceAll(path, ".", "_")
	}

	content, ok := webuiAssets[path]
	if !ok {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "fetchGeneratedAsset",
			"asset":    uri,
		}, "asset not found")
		return nil, "", fmt.Errorf("asset not found")
	}

	decoded, err := base64.StdEncoding.DecodeString(content.Data)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "fetchGeneratedAsset",
			"call":     "base64.StdEncoding.DecodeString",
			"asset":    uri,
		}, err.Error())
		return nil, "", fmt.Errorf("failed to decode asset")
	}

	return decoded, content.ContentType, nil
}

func (ui *WebUi) fetchDiscogsAsset(uri string) ([]byte, string, error) {
	imgFilename := ui.cfg.Discogs.CacheDir + "/" + path.Base(uri)

	content, err := ioutil.ReadFile(imgFilename)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "fetchDiscogsAsset",
			"call":     "ioutil.ReadFile",
			"asset":    uri,
		}, err.Error())
		return nil, "", fmt.Errorf("failed to fetch album art")
	}

	return content, "image/jpeg", nil
}

func (ui *WebUi) AssetHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	switch r.Method {
	case http.MethodGet:
		{
			content := []byte{}
			contentType := "text/plain"
			if r.URL.Path == "/" {
				content, contentType, err = ui.fetchGeneratedAsset(r.URL.Path)
			} else if strings.HasPrefix(r.URL.Path, "/css") {
				content, contentType, err = ui.fetchGeneratedAsset(r.URL.Path)
			} else if strings.HasPrefix(r.URL.Path, "/js") {
				content, contentType, err = ui.fetchGeneratedAsset(r.URL.Path)
			} else if strings.HasPrefix(r.URL.Path, "/img") {
				content, contentType, err = ui.fetchGeneratedAsset(r.URL.Path)
			} else if strings.HasPrefix(r.URL.Path, "/art") {
				content, contentType, err = ui.fetchDiscogsAsset(r.URL.Path)
			}

			if err != nil {
				log.Warningf(log.Fields{
					"package":  "webui",
					"function": "AssetHandler",
					"call":     "ui.fetchGeneratedAsset",
					"asset":    r.URL.Path,
				}, err.Error())
				http.Error(w, "", http.StatusInternalServerError)
				ui.accessLogEntry(r, 0, http.StatusInternalServerError)
				break
			}

			w.Header().Set("Content-Type", contentType)
			output := bytes.Buffer{}
			output.Write(content)

			w.Write(output.Bytes())
			ui.accessLogEntry(r, len(content), http.StatusOK)
			return
		}
	default:
		{
			http.Error(w, "", 405)
			ui.accessLogEntry(r, 0, 405)
			return
		}
	}
}

func (ui *WebUi) SocketHandler(w http.ResponseWriter, r *http.Request) {
	source := ui.getSource(r)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warningf(log.Fields{
			"package":  "webui",
			"function": "SocketHandler",
			"call":     "upgrader.Upgrade",
			"source":   source,
		}, err.Error())
		http.Error(w, "", 500)
		ui.accessLogEntry(r, 0, 500)
		return
	}

	for {
		if conn == nil {
			ui.wsAccessLogEntry(r, 500, "socket closed")
			break
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "webui",
				"function": "SocketHandler",
				"call":     "conn.ReadMessage",
				"source":   source,
			}, err.Error())
			conn = nil
			break
		}

		request := &wsClientRequest{}
		if err := json.Unmarshal(msg, request); err != nil {
			log.Warningf(log.Fields{
				"package":  "webui",
				"function": "SocketHandler",
				"call":     "json.Unmarshal",
				"source":   source,
			}, err.Error())
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
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.NewFailedResponse",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.wsSend",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
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
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.NewFailedResponse",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.wsSend",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
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
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.NewFailedResponse",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.wsSend",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
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
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.NewFailedResponse",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.wsSend",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
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
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.NewFailedResponse",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.wsSend",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
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
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.NewFailedResponse",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.wsSend",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
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
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.NewFailedResponse",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
					}

					err = ui.wsSend(conn, response)
					if err != nil {
						log.Warningf(log.Fields{
							"package":   "webui",
							"function":  "SocketHandler",
							"call":      "ui.wsSend",
							"operation": request.Operation,
							"source":    source,
						}, err.Error())
						break
					}
				}
			}
		default:
			{
				conn.Close()
				log.Warningf(log.Fields{
					"package":   "webui",
					"function":  "SocketHandler",
					"operation": request.Operation,
					"source":    source,
				}, "unknown operation")
				ui.wsAccessLogEntry(r, 500, fmt.Sprintf("%d", request.Operation))
				break
			}
		}
	}
}
