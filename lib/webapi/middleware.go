package webapi

import (
	"encoding/json"
	"fmt"

	"github.com/r3boot/go-musicbot/lib/mp3lib"
)

func (api *WebApi) newNowPlayingMsg() []byte {
	response := &NowPlayingResp{
		Data: NowPlaying{
			Title:    cache.Title,
			Duration: cache.Duration,
			Rating:   cache.Rating,
		},
		Pkt: "np_r",
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Warningf("WebApi.newNowPlayingMsg json.Marshal: %v", err)
		return nil
	}

	return data
}

func (api *WebApi) PlayQueueResponse() []byte {
	playQueue, err := api.mpd.GetPlayQueue()
	if err != nil {
		log.Warningf("WebApi.PlayQueueResponse: %v", err)
		return nil
	}

	response := GetQueueResp{
		Data: GetQueueRespData{
			Entries: playQueue,
			Size:    len(playQueue),
		},
		Pkt: "queue_r",
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Warningf("WebApi.PlayQueueResponse json.Marshal: %v", err)
		return nil
	}

	return data
}

func (api *WebApi) NowPlayingResponse() []byte {
	return api.newNowPlayingMsg()
}

func (api *WebApi) NextResponse() []byte {
	api.mpd.Next()
	return api.newNowPlayingMsg()
}

func (api *WebApi) BooResponse() []byte {
	fileName := api.mpd.NowPlaying()
	fullPath := api.yt.MusicDir + "/" + fileName
	newRating := api.mp3.DecreaseRating(fullPath)

	log.Infof("Rating for %s is now %d", fileName, newRating)

	if newRating == mp3lib.RATING_ZERO {
		api.mpd.Next()
		api.mp3.RemoveFile(fileName)

		response := fmt.Sprintf("Rating for %s is so low, it has been removed from the playlist", fileName[:len(fileName)-16])
		log.Infof("%s", response)
	}

	return api.newNowPlayingMsg()
}

func (api *WebApi) TuneResponse() []byte {
	fileName := api.mpd.NowPlaying()
	fullPath := api.yt.MusicDir + "/" + fileName
	newRating := api.mp3.IncreaseRating(fullPath)
	log.Infof("Rating for %s is now %d", fileName, newRating)
	return api.newNowPlayingMsg()
}
