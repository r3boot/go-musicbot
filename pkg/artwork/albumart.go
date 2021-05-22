package artwork

import (
	"github.com/r3boot/go-musicbot/pkg/mq"
	"github.com/sirupsen/logrus"
)

const (
	ModuleName = "AlbumArt"
)

type AlbumArt struct {
	sendChan chan mq.Message
	recvChan chan mq.Message
	quit     chan bool
	log      *logrus.Entry
}

func NewAlbumArt() (*AlbumArt, error) {
	art := &AlbumArt{
		sendChan: make(chan mq.Message, mq.MaxInFlight),
		recvChan: make(chan mq.Message, mq.MaxInFlight),
		quit:     make(chan bool),
		log:      logrus.WithFields(logrus.Fields{"caller": ModuleName}),
	}

	go art.MessagePipe()

	return art, nil
}

func (art *AlbumArt) GetSendChan() chan mq.Message {
	return art.sendChan
}

func (art *AlbumArt) GetRecvChan() chan mq.Message {
	return art.recvChan
}

func (art *AlbumArt) MessagePipe() {
	art.log.Debug("Starting MessagePipe")
	for {
		select {
		case msg := <-art.recvChan:
			{
				msgType := msg.GetMsgType()
				switch msgType {
				case mq.MsgNowPlaying:
					{
						npMsg := msg.GetContent().(*mq.PlaylistEntry)
						art.update(npMsg.Filename)
					}
				}
			}
		case <-art.quit:
			return
		}
	}
}

func (art *AlbumArt) update(fname string) {
	art.log.WithFields(logrus.Fields{
		"fname": fname,
	}).Debug("Fetching art")

	image := "image.jpeg"
	art.log.WithFields(logrus.Fields{
		"fname": fname,
		"image": image,
	}).Debug("Updated art")
	updateAAMsg := mq.NewUpdateAlbumArtMessage(fname, image)
	msg := mq.NewMessage(ModuleName, mq.MsgUpdateAlbumArt, updateAAMsg)
	art.sendChan <- *msg
}
