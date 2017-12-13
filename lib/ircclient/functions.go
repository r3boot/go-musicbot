package ircclient

import (
	"math/rand"
)

func (c *IrcClient) randomRadioMessage() string {
	n := rand.Int() % len(c.config.Bot.RadioMsgs)
	return c.config.Bot.RadioMsgs[n]
}

func (c *IrcClient) randomCh00nMessage() string {
	n := rand.Int() % len(c.config.Bot.Ch00nMsgs)
	return c.config.Bot.Ch00nMsgs[n]
}
