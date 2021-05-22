package mq

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Sender struct {
	name    string
	quit    chan bool
	channel chan Message
	log     *logrus.Entry
}

func (mq *MessageQueue) NewSender(name string, channel chan Message) error {
	if mq.numSenders == MaxSenders {
		return fmt.Errorf("MessageQueue.NewSender: len(mq.senders) == MaxSenders")
	}

	snd := &Sender{
		name:    name,
		channel: channel,
		quit:    make(chan bool),
		log: logrus.WithFields(logrus.Fields{
			"caller": ModuleName,
			"sender": name,
		}),
	}

	mq.senders = append(mq.senders, snd)

	go snd.MessagePipe(mq.transport)

	mq.numSenders += 1

	return nil
}

func (s Sender) MessagePipe(transport chan Message) {
	s.log.Debug("Starting MessagePipe")
	for {
		select {
		case msg := <-s.channel:
			{
				transport <- msg
			}
		case <-s.quit:
			return
		}
	}
}
