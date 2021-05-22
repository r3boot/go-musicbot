package mq

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Receiver struct {
	name    string
	channel chan Message
	log     *logrus.Entry
}

func (mq *MessageQueue) NewReceiver(name string, channel chan Message) error {
	if mq.numReceivers == MaxReceivers {
		return fmt.Errorf("MessageQueue.NewReceiver: len(mq.receivers) == MaxReceivers")
	}

	rcv := &Receiver{
		name:    name,
		channel: channel,
		log: logrus.WithFields(logrus.Fields{
			"caller":   ModuleName,
			"receiver": name,
		}),
	}

	mq.receivers = append(mq.receivers, rcv)

	mq.numReceivers += 1

	return nil
}
