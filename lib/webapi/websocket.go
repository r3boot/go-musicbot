package webapi

import (
	"encoding/json"
	"net/http"

	"fmt"

	"net/url"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (r WSServerResponse) ToJSON() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		return []byte("{\"i\":r.Id,\"o\":r.Operation,\"s\":false,\"m\":\"failed to encode to json\"}")
	}

	return data
}

func (a *WebAPI) HandleGetPlaylist(r *http.Request, conn *websocket.Conn, msgType int, request *WSClientRequest) {
	response := &WSServerResponse{
		ClientId:  request.Id,
		Operation: request.Operation,
	}

	data := a.Playlist.ToArray()
	response.Status = true
	response.Data = data
	conn.WriteMessage(msgType, response.ToJSON())
}

func (a *WebAPI) HandleGetArtists(r *http.Request, conn *websocket.Conn, msgType int, request *WSClientRequest) {
	response := &WSServerResponse{
		ClientId:  request.Id,
		Operation: request.Operation,
	}

	response.Status = true
	response.Data = a.Artists
	conn.WriteMessage(msgType, response.ToJSON())
}

func (a *WebAPI) HandleNowPlaying(r *http.Request, conn *websocket.Conn, msgType int, request *WSClientRequest) {
	response := &WSServerResponse{
		ClientId:  request.Id,
		Operation: request.Operation,
	}

	npData := a.mpdClient.NowPlaying()

	response.Status = true
	response.Data = npData

	conn.WriteMessage(msgType, response.ToJSON())
}

func (a *WebAPI) HandleNext(r *http.Request, conn *websocket.Conn, msgType int, request *WSClientRequest) {
	response := &WSServerResponse{
		ClientId:  request.Id,
		Operation: request.Operation,
	}

	nextTrack := a.mpdClient.Next()
	response.Status = true
	response.Data = nextTrack
	conn.WriteMessage(msgType, response.ToJSON())
	msg := fmt.Sprintf("Skipped to %s", nextTrack)
	wsOkResponse(r, "Next", msg)
}

func (a *WebAPI) HandleBoo(r *http.Request, conn *websocket.Conn, msgType int, request *WSClientRequest) {
	response := &WSServerResponse{
		ClientId:  request.Id,
		Operation: request.Operation,
	}

	npData := a.mpdClient.NowPlaying()
	newRating, err := a.id3Tags.DecreaseRating(npData.Filename)
	if err != nil {
		log.Warningf("WebAPI.HandleBoo: %v", err)
		response.Status = false
		response.Message = "failed to decrease rating"
		conn.WriteMessage(msgType, response.ToJSON())
	}
	msg := fmt.Sprintf("Rating for %s set to %d", npData.Title, newRating)
	wsOkResponse(r, "Boo", msg)
}

func (a *WebAPI) HandleTune(r *http.Request, conn *websocket.Conn, msgType int, request *WSClientRequest) {
	response := &WSServerResponse{
		ClientId:  request.Id,
		Operation: request.Operation,
	}

	npData := a.mpdClient.NowPlaying()
	newRating, err := a.id3Tags.IncreaseRating(npData.Filename)
	if err != nil {
		log.Warningf("WebAPI.HandleTune: %v", err)
		response.Status = false
		response.Message = "failed to increase rating"
		conn.WriteMessage(msgType, response.ToJSON())
		wsErrorResponse(r, "Tune", "failed to increase rating")
		return
	}
	msg := fmt.Sprintf("Rating for %s set to %d", npData.Title, newRating)
	wsOkResponse(r, "Tune", msg)
}

func (a *WebAPI) HandleRequest(r *http.Request, conn *websocket.Conn, msgType int, request *WSClientRequest) {
	response := &WSServerResponse{
		ClientId:  request.Id,
		Operation: request.Operation,
	}

	query, err := url.PathUnescape(request.Data.(string))
	if err != nil {
		log.Warningf("WebAPI.HandleRequest url.PathUnescape: %v", err)
		response.Status = false
		response.Message = "failed to decode data"
		conn.WriteMessage(msgType, response.ToJSON())
		wsErrorResponse(r, "Request", "failed to decode data")
		return
	}

	log.Debugf("WebAPI.HandleRequest: searching for %s", query)

	entry, err := a.mpdClient.Enqueue(query)
	if err != nil {
		log.Debugf("WebAPI.HandleRequest: %v", err)
		response.Status = false
		response.Message = "failed to submit request"
		conn.WriteMessage(msgType, response.ToJSON())
		wsErrorResponse(r, "Request", "failed to submit request")
		return
	}

	log.Debugf("WebAPI.HandleRequest: enqueued with id %d", entry.QPrio)

	response.Status = true
	response.Data = entry
	conn.WriteMessage(msgType, response.ToJSON())
	wsOkResponse(r, "Request", "submitted request to queue")
}

func (a *WebAPI) SocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warningf("WebApi.SocketHandler upgrader.Upgrade: %v", err)
		wsErrorResponse(r, "upgrade", "failed to upgrade socket")
		return
	}
	defer conn.Close()
	for {
		if conn == nil {
			log.Warningf("WebApi.SocketHandler: socket closed")
			wsErrorResponse(r, "socket", "closed")
			break
		}

		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Warningf("WebApi.SocketHandler conn.ReadMessage: %v", err)
			wsErrorResponse(r, "socket", "failed to read message")
			continue
		}

		request := &WSClientRequest{}
		if err := json.Unmarshal(msg, request); err != nil {
			log.Warningf("WebApi.SocketHandler json.Unmarshal: %v", err)
			log.Debugf("%v", string(msg))
			wsErrorResponse(r, "receive", "failed to unmarshal json")
			continue
		}

		switch request.Operation {
		case WS_GET_PLAYLIST:
			{
				a.HandleGetPlaylist(r, conn, msgType, request)
			}
		case WS_GET_ARTISTS:
			{
				a.HandleGetArtists(r, conn, msgType, request)
			}
		case WS_NEXT:
			{
				a.HandleNext(r, conn, msgType, request)
			}
		case WS_BOO:
			{
				a.HandleBoo(r, conn, msgType, request)
			}
		case WS_TUNE:
			{
				a.HandleTune(r, conn, msgType, request)
			}
		case WS_NOWPLAYING:
			{
				a.HandleNowPlaying(r, conn, msgType, request)
			}
		case WS_REQUEST:
			{
				a.HandleRequest(r, conn, msgType, request)
			}
		default:
			{
				conn.Close()
				log.Warningf("WebApi.SocketHandler: unknown operation received: %d", request.Operation)
				wsErrorResponse(r, "request", "unknown operation")
				break
			}
		}
	}
}
