package mq

import (
	"time"
)

const (
	MsgPlay = iota
	MsgNext
	MsgRewind
	MsgRequest
	MsgQueue
	MsgDownload
	MsgPlaylistAdd
	MsgUpdateIndex
	MsgGetNowPlaying
	MsgNowPlaying
	MsgIncreaseRating
	MsgDecreaseRating
	MsgUpdateAlbumArt
	MsgUpdateIDs
	MsgSongTooLong
	MsgAddToDB
	MsgUpdatePlaylist
	MsgDownloadAccepted
	MsgDownloadCompleted
	MsgSearchError
	MsgSearchResult
	MsgQueueError
	MsgQueueResult
	MsgGetQueueRequest
	MsgGetQueueResponse
)

type PlayMessage struct {
	Pos int
}

type NextMessage struct{}

type RewindMessage struct{}

type RequestMessage struct {
	Query     string
	Submitter string
}

type QueueMessage struct {
	Filename  string
	Id        int
	Submitter string
}

type DownloadMessage struct {
	Site      int
	ID        string
	Submitter string
}

type PlaylistAddMessage struct {
	Filename  string
	Submitter string
}

type UpdateIndexMessage struct {
	Filename  string
	Pos       int
	Rating    int
	Submitter string
}

type GetNowPlayingMessage struct{}

type NowPlayingMessage struct {
	Filename  string
	Pos       int
	Id        int
	Rating    int
	Duration  time.Time
	Submitter string
}

type IncreaseRatingMessage struct {
	Filename string
}

type DecreaseRatingMessage struct {
	Filename string
}

type UpdateAlbumArtMessage struct {
	Filename string
	Image    string
}

type UpdateIDsMessage struct {
	IDs []string
}

type SongTooLongMessage struct {
	ID        string
	Submitter string
}

type AddToDBMessage struct {
	Filename string
}

type PlaylistEntry struct {
	Filename     string
	LastModified time.Time
	Rating       int
	Duration     int
	Pos          int
	Id           int
	QPrio        int
	Submitter    string
}

type UpdatePlaylistMessage struct {
	Entries map[string]PlaylistEntry
}

type DownloadAcceptedMessage struct {
	ID        string
	Submitter string
}

type DownloadCompletedMessage struct {
	Filename  string
	Submitter string
}

type SearchErrorMessage struct {
	Error     string
	Submitter string
}

type SearchResultMessage struct {
	Filename  string
	Submitter string
}

type QueueErrorMessage struct {
	Error     string
	Submitter string
}

type QueueResultMessage struct {
	Filename  string
	Prio      int
	Submitter string
}

type GetQueueRequestMessage struct {
	Submitter string
}

type GetQueueResponseMessage struct {
	Queue     map[int]PlaylistEntry
	Submitter string
}

var (
	MsgTypeToString = map[int]string{
		MsgPlay:              "MsgPlay",
		MsgNext:              "MsgNext",
		MsgRewind:            "MsgRewind",
		MsgRequest:           "MsgRequest",
		MsgQueue:             "MsgQueue",
		MsgDownload:          "MsgDownload",
		MsgPlaylistAdd:       "MsgPlaylistAdd",
		MsgUpdateIndex:       "MsgUpdateIndex",
		MsgGetNowPlaying:     "MsgGetNowPlaying",
		MsgNowPlaying:        "MsgNowPlaying",
		MsgIncreaseRating:    "MsgIncreaseRating",
		MsgDecreaseRating:    "MsgDecreaseRating",
		MsgUpdateAlbumArt:    "MsgUpdateAlbumArt",
		MsgUpdateIDs:         "MsgUpdateIDs",
		MsgSongTooLong:       "MsgSongTooLong",
		MsgAddToDB:           "MsgAddToDB",
		MsgUpdatePlaylist:    "MsgUpdatePlaylist",
		MsgDownloadAccepted:  "MsgDownloadAccepted",
		MsgDownloadCompleted: "MsgDownloadCompleted",
		MsgSearchError:       "MsgSearchError",
		MsgSearchResult:      "MsgSearchResult",
		MsgQueueError:        "MsgQueueError",
		MsgQueueResult:       "MsgQueueResult",
		MsgGetQueueRequest:   "MsgGetQueueRequest",
		MsgGetQueueResponse:  "MsgGetQueueResponse",
	}
)

func NewPlayMessage(pos int) *PlayMessage {
	return &PlayMessage{
		Pos: pos,
	}
}

func NewNextMessage() *NextMessage {
	return &NextMessage{}
}

func NewRewindMessage() *RewindMessage {
	return &RewindMessage{}
}

func NewRequestMessage(query, submitter string) *RequestMessage {
	return &RequestMessage{
		Query:     query,
		Submitter: submitter,
	}
}

func NewQueueMessage(id int, fname, submitter string) *QueueMessage {
	return &QueueMessage{
		Id:        id,
		Filename:  fname,
		Submitter: submitter,
	}
}

func NewDownloadMessage(site int, yid, submitter string) *DownloadMessage {
	return &DownloadMessage{
		Site:      site,
		ID:        yid,
		Submitter: submitter,
	}
}

func NewPlaylistAddMessage(fname, submitter string) *PlaylistAddMessage {
	return &PlaylistAddMessage{
		Filename:  fname,
		Submitter: submitter,
	}
}

func NewUpdateIndexMessage(fname string, pos int) *UpdateIndexMessage {
	return &UpdateIndexMessage{
		Filename: fname,
		Pos:      pos,
	}
}

func NewGetNowPlayingMessage() *GetNowPlayingMessage {
	return &GetNowPlayingMessage{}
}

func NewNowPlayingMessage(fname string, pos, id, rating int, duration time.Time, submitter string) *NowPlayingMessage {
	return &NowPlayingMessage{
		Filename:  fname,
		Pos:       pos,
		Id:        id,
		Rating:    rating,
		Duration:  duration,
		Submitter: submitter,
	}
}

func NewIncreaseRatingMessage(fname string) *IncreaseRatingMessage {
	return &IncreaseRatingMessage{
		Filename: fname,
	}
}

func NewDecreaseRatingMessage(fname string) *DecreaseRatingMessage {
	return &DecreaseRatingMessage{
		Filename: fname,
	}
}

func NewUpdateAlbumArtMessage(fname, image string) *UpdateAlbumArtMessage {
	return &UpdateAlbumArtMessage{
		Filename: fname,
		Image:    image,
	}
}

func NewUpdateIDsMessage(idList []string) *UpdateIDsMessage {
	return &UpdateIDsMessage{
		IDs: idList,
	}
}

func NewSongTooLongMessage(id, submitter string) *SongTooLongMessage {
	return &SongTooLongMessage{
		ID:        id,
		Submitter: submitter,
	}
}

func NewAddToDBMessage(fname string) *AddToDBMessage {
	return &AddToDBMessage{
		Filename: fname,
	}
}

func NewUpdatePlaylistMessage(playlist map[string]PlaylistEntry) *UpdatePlaylistMessage {
	return &UpdatePlaylistMessage{
		Entries: playlist,
	}
}

func NewDownloadAcceptedMessage(id, submitter string) *DownloadAcceptedMessage {
	return &DownloadAcceptedMessage{
		ID:        id,
		Submitter: submitter,
	}
}

func NewDownloadCompletedMessage(fname, submitter string) *DownloadCompletedMessage {
	return &DownloadCompletedMessage{
		Filename:  fname,
		Submitter: submitter,
	}
}

func NewSearchErrorMessage(error, submitter string) *SearchErrorMessage {
	return &SearchErrorMessage{
		Error:     error,
		Submitter: submitter,
	}
}

func NewSearchResultMessage(fname, submitter string) *SearchResultMessage {
	return &SearchResultMessage{
		Filename:  fname,
		Submitter: submitter,
	}
}

func NewQueueErrorMessage(error, submitter string) *QueueErrorMessage {
	return &QueueErrorMessage{
		Error:     error,
		Submitter: submitter,
	}
}

func NewQueueResultMessage(fname, submitter string, prio int) *QueueResultMessage {
	return &QueueResultMessage{
		Filename:  fname,
		Prio:      prio,
		Submitter: submitter,
	}
}

func NewGetQueueRequestMessage(submitter string) *GetQueueRequestMessage {
	return &GetQueueRequestMessage{
		Submitter: submitter,
	}
}

func NewGetQueueResponseMessage(queue map[int]PlaylistEntry, submitter string) *GetQueueResponseMessage {
	return &GetQueueResponseMessage{
		Queue:     queue,
		Submitter: submitter,
	}
}
