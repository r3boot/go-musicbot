package rating

import (
	"github.com/r3boot/go-musicbot/pkg/mq"
	"github.com/sirupsen/logrus"
)

const (
	ModuleName = "Ratings"
)

type Ratings struct {
	recvChan chan mq.Message
	sendChan chan mq.Message
	quit     chan bool
	log      *logrus.Entry
}

func NewRatings() (*Ratings, error) {
	ratings := &Ratings{
		recvChan: make(chan mq.Message, mq.MaxInFlight),
		sendChan: make(chan mq.Message, mq.MaxInFlight),
		quit:     make(chan bool),
		log:      logrus.WithFields(logrus.Fields{"caller": ModuleName}),
	}

	go ratings.MessagePipe()

	return ratings, nil
}

func (ratings *Ratings) GetRecvChan() chan mq.Message {
	return ratings.recvChan
}

func (ratings *Ratings) GetSendChan() chan mq.Message {
	return ratings.sendChan
}

func (ratings *Ratings) MessagePipe() {
	ratings.log.Debug("Starting MessagePipe")
	for {
		select {
		case msg := <-ratings.recvChan:
			{
				msgType := msg.GetMsgType()
				switch msgType {
				case mq.MsgIncreaseRating:
					{
						incRatingMsg := msg.GetContent().(*mq.IncreaseRatingMessage)
						ratings.increaseRating(incRatingMsg.Filename)
					}
				case mq.MsgDecreaseRating:
					{
						decRatingMsg := msg.GetContent().(*mq.DecreaseRatingMessage)
						ratings.decreaseRating(decRatingMsg.Filename)
					}
				}
			}
		case <-ratings.quit:
			return
		}
	}
}

func (ratings *Ratings) increaseRating(fname string) {
	ratings.log.WithFields(logrus.Fields{
		"fname": fname,
	}).Info("Increasing rating")

	pos := 345
	rating := 42
	ratings.log.WithFields(logrus.Fields{
		"fname":  fname,
		"pos":    pos,
		"rating": rating,
	}).Debug("Sending UpdateIndex")

	updateIndexMsg := mq.NewUpdateIndexMessage(fname, pos)
	msg := mq.NewMessage(ModuleName, mq.MsgUpdateIndex, updateIndexMsg)
	ratings.sendChan <- *msg
}

func (ratings *Ratings) decreaseRating(fname string) {
	ratings.log.WithFields(logrus.Fields{
		"fname": fname,
	}).Info("Decreasing rating")

	pos := 678
	rating := 666
	ratings.log.WithFields(logrus.Fields{
		"fname":  fname,
		"pos":    pos,
		"rating": rating,
	}).Debug("Sending UpdateIndex")

	updateIndexMsg := mq.NewUpdateIndexMessage(fname, pos)
	msg := mq.NewMessage(ModuleName, mq.MsgUpdateIndex, updateIndexMsg)
	ratings.sendChan <- *msg
}
