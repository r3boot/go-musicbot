package webapi

import (
	"encoding/json"
	"fmt"

	"github.com/r3boot/go-musicbot/lib/id3tags"
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
	newRating, err := api.id3.DecreaseRating(fullPath)
	if err != nil {
		log.Warningf("WebApi.BooResponse: %v", err)
		return nil
	}

	log.Infof("Rating for %s is now %d", fileName, newRating)

	if newRating == id3tags.RATING_ZERO {
		api.mpd.Next()
		err = api.id3.RemoveFile(fileName)
		if err != nil {
			log.Warningf("WebApi.BooResponse: %v", err)
			return nil
		}

		response := fmt.Sprintf("Rating for %s is so low, it has been removed from the playlist", fileName[:len(fileName)-16])
		log.Infof("%s", response)
	}

	return api.newNowPlayingMsg()
}

func (api *WebApi) TuneResponse() []byte {
	fileName := api.mpd.NowPlaying()
	fullPath := api.yt.MusicDir + "/" + fileName
	newRating, err := api.id3.IncreaseRating(fullPath)
	if err != nil {
		log.Warningf("WebApi.TuneResponse: %v", err)
		return nil
	}

	log.Infof("Rating for %s is now %d", fileName, newRating)
	return api.newNowPlayingMsg()
}
