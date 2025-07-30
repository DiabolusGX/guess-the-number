package utils

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
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

type MessageRequest struct {
	ChannelID   snowflake.ID
	Content     string
	Emoji       Emoji
	IsEphemeral bool
	WithVote    bool
}

func EventReply(event *events.ApplicationCommandInteractionCreate, request MessageRequest) error {
	content := request.Content
	if request.Emoji != EmojiUnspecified {
		content = fmt.Sprintf("%s | %s", request.Emoji.String(), request.Content)
	}

	if request.WithVote {
		content = fmt.Sprintf("%s\n\n> *Show your support by voting for the bot ❤️ https://top.gg/bot/818420448131285012/vote*", content)
	}

	// Since interaction is deferred, we need to update the response instead of creating a new one
	messageUpdateRequest := discord.MessageUpdate{Content: &content}
	if request.IsEphemeral {
		// Ephemeral flag was already set during defer, no need to set it again
	}

	_, err := event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), messageUpdateRequest)
	return err
}

func SendMessage(rest rest.Rest, request MessageRequest) (*discord.Message, error) {
	content := request.Content
	if request.Emoji != EmojiUnspecified {
		content = fmt.Sprintf("%s | %s", request.Emoji.String(), request.Content)
	}

	if request.WithVote {
		content = fmt.Sprintf("%s\n\n> *Show your support by voting for the bot ❤️ https://top.gg/bot/818420448131285012/vote*", content)
	}

	messageCreateRequest := discord.MessageCreate{Content: content}
	if request.IsEphemeral {
		messageCreateRequest.Flags = discord.MessageFlagEphemeral
	}

	msg, err := rest.CreateMessage(request.ChannelID, messageCreateRequest)
	return msg, err
}

func SendDM(rest rest.Rest, userID snowflake.ID, request MessageRequest) error {
	// Create (or fetch) a DM channel with the user
	dmChannel, err := rest.CreateDMChannel(userID)
	if err != nil {
		return err
	}

	request.ChannelID = dmChannel.ID()
	_, err = SendMessage(rest, request)
	return err
}

// LogToChannel sends a formatted message to the configured log channel
func LogToChannel(rest rest.Rest, logChannelID string, title, description, footer string) error {
	if logChannelID == "" {
		return nil // No log channel configured, skip logging
	}

	channelID, err := snowflake.Parse(logChannelID)
	if err != nil {
		return err
	}

	embed := discord.Embed{
		Title:       title,
		Description: description,
		Color:       0x5865F2, // Discord Blurple color
	}

	if footer != "" {
		embed.Footer = &discord.EmbedFooter{
			Text: footer,
		}
	}

	_, err = rest.CreateMessage(channelID, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
	return err
}
