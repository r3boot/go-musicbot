package ircclient

import (
	"math/rand"
)

func (c *IrcClient) isValidCommand(cmd string) (string, bool) {
	result := RE_VALIDCMD.FindAllStringSubmatch(cmd, -1)

	if len(result) == 0 {
		log.Warningf("IrcClient.isValidCommand: Did not find any valid command", cmd)
		return "", false
	}

	wantedCommand := result[0][1]
	for _, validCommand := range c.config.Bot.ValidCommands {
		if wantedCommand == validCommand {
			return wantedCommand, true
		}
	}

	log.Warningf("IrcClient.isValidCommand: Unknown command: %s", cmd)
	return "", false
}

func (c *IrcClient) randomRadioMessage() string {
	n := rand.Int() % len(c.config.Bot.RadioMsgs)
	return c.config.Bot.RadioMsgs[n]
}
