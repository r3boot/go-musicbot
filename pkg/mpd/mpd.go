package mpd

import (
	"fmt"
	"github.com/r3boot/go-musicbot/pkg/config"
	"github.com/r3boot/go-musicbot/pkg/id3tags"
	"github.com/r3boot/go-musicbot/pkg/indexer"
	"github.com/r3boot/go-musicbot/pkg/mq"
	"github.com/sirupsen/logrus"
	"path"
	"strconv"
	"strings"

	gompd "github.com/r3boot/gompd/mpd"
	"time"
)

const (
	ModuleName = "mpd"

	MaxQueueSize = 8
)

type MpdClient struct {
	sendChan chan mq.Message
	recvChan chan mq.Message
	quit     chan bool
	log      *logrus.Entry
	cfg      *config.MusicBotConfig
	tags     *id3tags.ID3Tags
	search   *indexer.Search
	conn     *gompd.Client
}

func NewMpdClient(cfg *config.MusicBotConfig, tags *id3tags.ID3Tags, search *indexer.Search) (*MpdClient, error) {
	mpdClient := &MpdClient{
		sendChan: make(chan mq.Message, mq.MaxInFlight),
		recvChan: make(chan mq.Message, mq.MaxInFlight),
		log:      logrus.WithFields(logrus.Fields{"caller": ModuleName}),
		cfg:      cfg,
		tags:     tags,
		search:   search,
	}

	go mpdClient.Connect()
	go mpdClient.MessagePipe()

	return mpdClient, nil
}

func (client *MpdClient) GetRecvChan() chan mq.Message {
	return client.recvChan
}

func (client *MpdClient) GetSendChan() chan mq.Message {
	return client.sendChan
}

func (client *MpdClient) MessagePipe() {
	client.log.Debug("Starting MessagePipe")
	for {
		select {
		case msg := <-client.recvChan:
			{
				msgType := msg.GetMsgType()
				switch msgType {
				case mq.MsgPlay:
					{
						playMsg := msg.GetContent().(*mq.PlayMessage)
						client.play(playMsg.Pos)
					}
				case mq.MsgNext:
					{
						client.next()

					}
				case mq.MsgRewind:
					{
						client.rewind()
					}
				case mq.MsgQueue:
					{
						qMsg := msg.GetContent().(*mq.QueueMessage)
						client.addToQueue(qMsg.Id, qMsg.Filename, qMsg.Submitter)
					}
				case mq.MsgGetQueueRequest:
					{
						qrMsg := msg.GetContent().(*mq.GetQueueRequestMessage)
						client.sendQueueTo(qrMsg.Submitter)
					}
				case mq.MsgGetNowPlaying:
					{
						client.nowPlaying()
						client.getQueue()
					}
				case mq.MsgAddToDB:
					{
						atdbMsg := msg.GetContent().(*mq.AddToDBMessage)
						client.updateDB(atdbMsg.Filename)
					}
				default:
					{
						continue
					}
				}
			}
		case <-client.quit:
			{
				return
			}
		}
	}
}

func (client *MpdClient) Connect() {
	var err error

	address := fmt.Sprintf("%s:%d", client.cfg.MPD.Address, client.cfg.MPD.Port)

	if client.cfg.MPD.Password != "" {
		client.conn, err = gompd.DialAuthenticated("tcp", address, client.cfg.MPD.Password)
		if err != nil {
			client.log.WithFields(logrus.Fields{
				"address": address,
				"error":   err,
			}).Warn("Failed to connect to MPD")
			client.conn = nil
			return
		}
	} else {
		client.conn, err = gompd.Dial("tcp", address)
		if err != nil {
			client.log.WithFields(logrus.Fields{
				"address": address,
				"error":   err,
			}).Warn("Failed to connect to MPD")
			client.conn = nil
			return
		}
	}

	// Once connected, update the database
	client.updateDB("")

	// Ensure we stay connected to MPD
	go client.Keepalive()

	// Ensure that MPD is in random mode if we have enabled queueing
	if client.cfg.Features.Queue {
		client.conn.Random(true)
		client.log.Debug("Queueing enabled, force-enabling random mode")
	}
}

func (client *MpdClient) Keepalive() {
	for {
		time.Sleep(1 * time.Second)
		err := client.conn.Ping()
		if err != nil {
			client.log.WithFields(logrus.Fields{
				"error": err,
			}).Warn("Connection failed, reconnecting")
			break
		}
	}

	go client.Connect()
}

func (client *MpdClient) play(pos int) {
	client.log.WithFields(logrus.Fields{
		"pos": pos,
	}).Info("Playing track")
}

func (client *MpdClient) next() {
	client.log.Info("Playing next track")
}

func (client *MpdClient) rewind() {
	client.log.Info("Rewinding current track")
}

func (client *MpdClient) nowPlaying() {
	attrs, err := client.conn.CurrentSong()
	if err != nil {
		client.log.WithFields(logrus.Fields{
			"cmd":   "nowPlaying",
			"error": err,
		}).Warn("Failed to fetch current song info")
	}

	fname := attrs["file"]

	entry, err := client.search.FindByFilename(fname)
	if err != nil {
		client.log.WithFields(logrus.Fields{
			"cmd":   "nowPlaying",
			"error": err,
		}).Warn("Failed to search PlaylistEntry")
	}

	client.log.WithFields(logrus.Fields{
		"fname":     fname,
		"pos":       entry.Pos,
		"id":        entry.Id,
		"duration":  entry.Duration,
		"rating":    entry.Rating,
		"submitter": entry.Submitter,
	}).Debug("Sending NowPlaying")

	msg := mq.NewMessage(ModuleName, mq.MsgNowPlaying, entry)
	client.sendChan <- *msg
}

func (client *MpdClient) playlistInfo() map[string]mq.PlaylistEntry {
	if client.conn == nil {
		client.log.WithFields(logrus.Fields{
			"cmd": "playlistInfo",
		}).Warn("Not connected to MPD")
		return nil
	}

	attrs, err := client.conn.PlaylistInfo(-1, -1)
	if err != nil {
		client.log.WithFields(logrus.Fields{
			"cmd":   "playlistInfo",
			"error": err,
		}).Warn("Failed to fetch playlist")
		return nil
	}

	playlist := make(map[string]mq.PlaylistEntry, len(attrs))
	for _, entry := range attrs {
		fname := entry["file"]
		if !strings.HasSuffix(fname, ".mp3") {
			continue
		}
		lm, err := time.Parse("2006-01-02T15:04:5Z", entry["Last-Modified"])
		if err != nil {
			client.log.WithFields(logrus.Fields{
				"cmd":   "playlistInfo",
				"error": err,
				"fname": fname,
			}).Warn("Failed to parse timestamp")
			continue
		}

		rating, err := strconv.Atoi(entry["Track"])
		if err != nil {
			client.log.WithFields(logrus.Fields{
				"cmd":   "playlistInfo",
				"error": err,
				"fname": fname,
			}).Warn("Failed to parse rating")
			continue
		}

		duration, err := strconv.Atoi(entry["Time"])
		if err != nil {
			client.log.WithFields(logrus.Fields{
				"cmd":   "playlistInfo",
				"error": err,
				"fname": fname,
			}).Warn("Failed to parse Time")
			continue
		}

		pos, err := strconv.Atoi(entry["Pos"])
		if err != nil {
			client.log.WithFields(logrus.Fields{
				"cmd":   "playlistInfo",
				"error": err,
				"fname": fname,
			}).Warn("Failed to parse Pos")
			continue
		}

		id, err := strconv.Atoi(entry["Id"])
		if err != nil {
			client.log.WithFields(logrus.Fields{
				"cmd":   "playlistInfo",
				"error": err,
				"fname": fname,
			}).Warn("Failed to parse Id")
			continue
		}

		prio := -1
		tmp, ok := entry["Prio"]
		if ok {
			prio, err = strconv.Atoi(tmp)
			if err != nil {
				client.log.WithFields(logrus.Fields{
					"cmd":   "playlistInfo",
					"error": err,
					"fname": fname,
				}).Warn("Failed to parse Prio")
				continue
			}
		}

		playlist[fname] = mq.PlaylistEntry{
			Filename:     fname,
			LastModified: lm,
			Rating:       rating,
			Duration:     duration,
			Pos:          pos,
			Id:           id,
			QPrio:        prio,
		}
	}

	return playlist
}

func (client *MpdClient) getQueue() map[int]mq.PlaylistEntry {
	playlist := client.playlistInfo()

	queue := make(map[int]mq.PlaylistEntry)
	for _, entry := range playlist {
		if entry.QPrio != -1 {
			if _, ok := queue[entry.QPrio]; ok {
				client.log.WithFields(logrus.Fields{
					"error": "entry already exists",
				}).Warn("Unable to update queue")
			}
			queue[entry.QPrio] = entry
		}
	}

	return queue
}

func (client *MpdClient) addToQueue(id int, fname, submitter string) {
	qLog := client.log.WithFields(logrus.Fields{
		"id":        id,
		"fname":     fname,
		"submitter": submitter,
	})

	queue := client.getQueue()
	queueLen := len(queue)
	if queueLen > MaxQueueSize {
		qLog.Warn("queue is full")
		return
	}

	for id, entry := range queue {
		fmt.Printf("%d) %v\n", id, entry)
	}

	prio := queueLen

	err := client.conn.PrioId(id, prio)
	if err != nil {
		e := fmt.Sprintf("client.conn.PrioId: %v", err)
		qLog.WithFields(logrus.Fields{
			"error": e,
			"prio":  prio,
			"id":    id,
		}).Warn("Failed to set prio")

		qeMsg := mq.NewQueueErrorMessage(e, submitter)
		msg := mq.NewMessage(ModuleName, mq.MsgQueueError, qeMsg)
		client.sendChan <- *msg

		return
	}

	qLog.WithFields(logrus.Fields{
		"prio": prio,
	}).Info("Added track to queue")

	qrMsg := mq.NewQueueResultMessage(fname, submitter, MaxQueueSize-prio)
	msg := mq.NewMessage(ModuleName, mq.MsgQueueResult, qrMsg)
	client.sendChan <- *msg

	queue = client.getQueue()
	for id, entry := range queue {
		fmt.Printf("%d) %v\n", id, entry)
	}
}

func (client *MpdClient) sendQueueTo(submitter string) {
	queue := client.getQueue()
	respMsg := mq.NewGetQueueResponseMessage(queue, submitter)
	msg := mq.NewMessage(ModuleName, mq.MsgGetQueueResponse, respMsg)
	client.sendChan <- *msg
}

func (client *MpdClient) updateDB(fname string) {
	if client.conn == nil {
		client.log.WithFields(logrus.Fields{
			"cmd": "updateDB",
		}).Warn("Not connected to MPD")
		return
	}

	jobId := 0

	if fname != "" {
		fname = path.Base(fname)
	}

	jobId, err := client.conn.Update(fname)

	if err != nil {
		client.log.WithFields(logrus.Fields{
			"cmd":   "updateDB",
			"error": err,
		}).Warn("Failed to update DB")
		return
	}

	for {
		attrs, err := client.conn.Status()
		if err != nil {
			client.log.WithFields(logrus.Fields{
				"cmd":   "updateDB",
				"error": err,
			}).Warn("Failed to update DB")
			return
		}

		value, ok := attrs["updating_db"]
		if ok && value == strconv.Itoa(jobId) {
			time.Sleep(100 * time.Millisecond)
		}
		break
	}

	if fname != "" {
		client.log.WithFields(logrus.Fields{
			"cmd":   "updateDB",
			"fname": fname,
		}).Debug("Added to database")
	} else {
		client.log.WithFields(logrus.Fields{
			"cmd": "updateDB",
		}).Debug("Updated database")

	}

	playlist := client.playlistInfo()
	if playlist == nil {
		client.log.WithFields(logrus.Fields{
			"cmd":   "updateDB",
			"fname": fname,
		}).Warn("Failed to fetch playlist")
		return
	}

	submitters := client.tags.GetAllSubmitters()
	if submitters == nil {
		client.log.WithFields(logrus.Fields{
			"cmd":   "updateDB",
			"fname": fname,
		}).Warn("Failed to fetch submitters")
		return
	}

	if fname != "" {
		track := make(map[string]mq.PlaylistEntry)
		track[fname] = playlist[fname]
		upMsg := mq.NewUpdatePlaylistMessage(track)
		msg := mq.NewMessage(ModuleName, mq.MsgUpdatePlaylist, upMsg)
		client.sendChan <- *msg
	} else {
		upMsg := mq.NewUpdatePlaylistMessage(playlist)
		msg := mq.NewMessage(ModuleName, mq.MsgUpdatePlaylist, upMsg)
		client.sendChan <- *msg
	}

}
