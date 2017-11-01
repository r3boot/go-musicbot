package webapi

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/r3boot/go-musicbot/lib/mp3lib"
)

func (api *WebApi) newNowPlayingMsg() []byte {
	response := &NowPlayingResp{
		Data: NowPlaying{
			Title:    gTitle,
			Duration: gDuration,
			Rating:   gRating,
		},
		Pkt: "np_r",
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal response: %v\n", err)
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
	fmt.Printf("IrcClient.HandleDecreaseRating rating for %s is now %d\n", fileName, newRating)
	if newRating == mp3lib.RATING_ZERO {
		api.mpd.Next()
		api.mp3.RemoveFile(fileName)
		response := fmt.Sprintf("Rating for %s is so low, it has been removed from the playlist", fileName[:len(fileName)-16])
		fmt.Printf("%s\n", response)
	} else {
		response := fmt.Sprintf("Rating for %s is %d/10 .. BOOO!!!!", fileName[:len(fileName)-16], newRating)
		fmt.Printf("%s\n", response)
	}
	return api.newNowPlayingMsg()
}

func (api *WebApi) TuneResponse() []byte {
	fileName := api.mpd.NowPlaying()
	fullPath := api.yt.MusicDir + "/" + fileName
	newRating := api.mp3.IncreaseRating(fullPath)
	fmt.Printf("IrcClient.HandleIncreaseRating rating for %s is now %d\n", fileName, newRating)
	return api.newNowPlayingMsg()
}
