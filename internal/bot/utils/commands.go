package utils

import (
	"fmt"

	"github.com/disgoorg/snowflake/v2"
)

type CommandName string

const (
	defaultBotClientID = snowflake.ID(818420448131285012)
	testBotClientID    = snowflake.ID(686988635999830076)

	CommandStart     CommandName = "start"
	CommandGameStats CommandName = "game stats"
)

var applicationCommandsByBotID = map[snowflake.ID]map[CommandName]string{
	defaultBotClientID: {
		CommandStart:     "",
		CommandGameStats: "",
	},
	testBotClientID: {
		CommandStart:     "1397690638379651097",
		CommandGameStats: "1401599456654266389",
	},
}

func MentionApplicationCommand(botID snowflake.ID, commandName CommandName) string {
	return fmt.Sprintf("</%s:%s>", commandName, applicationCommandsByBotID[botID][commandName])
}
