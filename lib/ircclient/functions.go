package ircclient

import (
	"math/rand"
)

func (c *IrcClient) randomRadioMessage() string {
	n := rand.Int() % len(c.config.Bot.RadioMsgs)
	return c.config.Bot.RadioMsgs[n]
}
