package ircclient

import (
	"fmt"
	"math/rand"
	"regexp"
)

func (c *IrcClient) isValidCommand(cmd string) (string, bool) {
	cmdReString := fmt.Sprintf("^\\%s([a-z\\+\\-]{2,8})", c.config.Bot.CommandChar)
	reValidCmd := regexp.MustCompile(cmdReString)

	result := reValidCmd.FindAllStringSubmatch(cmd, -1)

	if len(result) == 0 {
		fmt.Printf("isValidCommand: %s does not match the valid command regexp\n", cmd)
		return "", false
	}

	wantedCommand := result[0][1]
	for _, validCommand := range c.config.Bot.ValidCommands {
		if wantedCommand == validCommand {
			return wantedCommand, true
		}
	}

	fmt.Printf("isValidCommand: Unknown command: %s\n", cmd)
	return "", false
}

func (c *IrcClient) randomRadioMessage() string {
	n := rand.Int() % len(c.config.Bot.RadioMsgs)
	return c.config.Bot.RadioMsgs[n]
}
