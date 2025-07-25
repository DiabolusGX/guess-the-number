package utils

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

func SendEmbed(rest rest.Rest, channelID snowflake.ID, title, description string, color int) error {
	_, err := rest.CreateMessage(channelID, discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Title:       title,
				Description: description,
				Color:       color,
			},
		},
	})
	return err
}
