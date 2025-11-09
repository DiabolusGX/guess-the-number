package utils

import (
	"fmt"

	"github.com/disgoorg/snowflake/v2"
)

type CommandName string

const (
	defaultBotClientID = snowflake.ID(818420448131285012)
	testBotClientID    = snowflake.ID(686988635999830076)

	CommandHelp      CommandName = "help"
	CommandStart     CommandName = "start"
	CommandGameStats CommandName = "game stats"
)

var applicationCommandsByBotID = map[snowflake.ID]map[CommandName]string{
	defaultBotClientID: {
		CommandHelp:      "1403827377288511508",
		CommandStart:     "1403827377288511515",
		CommandGameStats: "1403827377288511516",
	},
	testBotClientID: {
		CommandHelp:      "1403532388365242368",
		CommandStart:     "1397690638379651097",
		CommandGameStats: "1401599456654266389",
	},
}

func MentionApplicationCommand(botID snowflake.ID, commandName CommandName) string {
	return fmt.Sprintf("</%s:%s>", commandName, applicationCommandsByBotID[botID][commandName])
}
