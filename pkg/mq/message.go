package mq

import (
	"github.com/sirupsen/logrus"
	"time"
)

type Message struct {
	sender  string
	ts      time.Time
	msgtype int
	content interface{}
}

func NewMessage(sender string, msgType int, content interface{}) *Message {
	msg := &Message{
		sender:  sender,
		ts:      time.Now(),
		msgtype: msgType,
		content: content,
	}

	return msg
}

func (msg *Message) GetMsgType() int {
	return msg.msgtype
}

func (msg *Message) GetContent() interface{} {
	return msg.content
}

func (msg *Message) Dump(message string) {
	logrus.WithFields(logrus.Fields{
		"caller":  ModuleName,
		"sender":  msg.sender,
		"ts":      msg.ts,
		"msgtype": MsgTypeToString[msg.msgtype],
	}).Debug(message)
}
