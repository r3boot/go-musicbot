package mq

import (
	"github.com/sirupsen/logrus"
)

const (
	ModuleName = "MessageQueue"

	MaxSenders   = 64
	MaxReceivers = 64
	MaxInFlight  = 8
)

type MessageQueue struct {
	senders      []*Sender
	receivers    []*Receiver
	numSenders   int
	numReceivers int
	quit         chan bool
	log          *logrus.Entry
	transport    chan Message
}

func NewMessageQueue() (*MessageQueue, error) {
	mq := &MessageQueue{
		senders:   make([]*Sender, MaxSenders),
		receivers: make([]*Receiver, MaxReceivers),
		transport: make(chan Message, MaxInFlight),
		log:       logrus.WithFields(logrus.Fields{"caller": ModuleName}),
	}

	go mq.ReceiverDemux()

	return mq, nil
}

func (mq *MessageQueue) ReceiverDemux() {
	mq.log.Debug("Starting ReceiverDemux")
	for {
		select {
		case msg := <-mq.transport:
			{
				for _, receiver := range mq.receivers {
					if receiver == nil {
						continue
					}
					receiver.channel <- msg
				}
			}
		case <-mq.quit:
			{
				return
			}
		}
	}
}
