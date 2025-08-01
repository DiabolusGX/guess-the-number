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
	IsEphemeral bool

	Content string
	Emoji   Emoji

	UseEmbed         bool
	EmbedTitle       string
	EmbedDescription string
	EmbedColor       int
	Thumbnail        string

	WithVote          bool
	WithSupportServer bool
}

func EventReply(event *events.ApplicationCommandInteractionCreate, request MessageRequest) error {
	if request.ChannelID == 0 {
		request.ChannelID = event.Channel().ID()
	}

	// send a non-ephemeral response by deleting the ephemeral deferred response and sending a new message to the channel
	if !request.IsEphemeral {
		err := event.Client().Rest().DeleteInteractionResponse(event.ApplicationID(), event.Token())
		if err != nil {
			return err
		}
		_, err = SendMessage(event.Client().Rest(), request)
		return err
	}

	var messageUpdateRequest discord.MessageUpdate

	components := buildComponents(request)

	if request.UseEmbed {
		// Create embed and components
		embeds := buildEmbeds(request)
		messageUpdateRequest = discord.MessageUpdate{
			Embeds:     &embeds,
			Components: &components,
		}
	} else {
		// Original content-based logic
		content := request.Content
		if request.Emoji != EmojiUnspecified {
			content = fmt.Sprintf(emojiContentFormat, request.Emoji.String(), request.Content)
		}

		if request.WithVote {
			content = fmt.Sprintf("%s\n\n> *Show your support by voting for the bot ❤️ %s*", content, voteLink)
		}

		if request.WithSupportServer {
			content = fmt.Sprintf("%s\n\n> *Need help? Join our support server: %s*", content, supportServerLink)
		}

		messageUpdateRequest = discord.MessageUpdate{
			Content:    &content,
			Components: &components,
		}
	}

	_, err := event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), messageUpdateRequest)
	return err
}

func SendMessage(rest rest.Rest, request MessageRequest) (*discord.Message, error) {
	var messageCreateRequest discord.MessageCreate

	components := buildComponents(request)

	if request.UseEmbed {
		// Create embed and components
		embeds := buildEmbeds(request)
		messageCreateRequest = discord.MessageCreate{
			Embeds:     embeds,
			Components: components,
		}
	} else {
		// Original content-based logic
		content := request.Content
		if request.Emoji != EmojiUnspecified {
			content = fmt.Sprintf(emojiContentFormat, request.Emoji.String(), request.Content)
		}

		if request.WithVote {
			content = fmt.Sprintf("%s\n\n> *Show your support by voting for the bot ❤️ https://top.gg/bot/818420448131285012/vote*", content)
		}

		if request.WithSupportServer {
			content = fmt.Sprintf("%s\n\n> *Need help? Join our support server: %s*", content, supportServerLink)
		}

		messageCreateRequest = discord.MessageCreate{
			Content:    content,
			Components: components,
		}
	}

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
		Color:       discordBlurple,
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

// buildEmbeds creates a Discord embed from a MessageRequest
func buildEmbeds(request MessageRequest) []discord.Embed {
	embedColor := defaultEmbedColor
	if request.EmbedColor != 0 {
		embedColor = request.EmbedColor
	}

	embed := discord.Embed{
		Color:     embedColor,
		Thumbnail: &discord.EmbedResource{URL: botAvatarURL},
	}

	// Set title and description
	if request.EmbedTitle != "" {
		embed.Title = request.EmbedTitle
	}

	// Handle content based on whether we have description or content
	description := request.EmbedDescription
	if description == "" && request.Content != "" {
		description = request.Content
		if request.Emoji != EmojiUnspecified {
			description = fmt.Sprintf(emojiContentFormat, request.Emoji.String(), description)
		}
	}
	embed.Description = description

	// Prepare footer
	var footerText string
	var hasFooter bool

	// if request.WithVote {
	// 	footerText = "Show your support by voting for the bot ❤️"
	// 	hasFooter = true
	// }

	// if request.WithSupportServer {
	// 	footerText = "Need help? Join our support server!"
	// 	hasFooter = true
	// }

	if footerText != "" {
		embed.Footer = &discord.EmbedFooter{Text: footerText}
		hasFooter = true
	}

	if !hasFooter {
		embed.Footer = &discord.EmbedFooter{
			Text:    "Made with ❤️ by DiabolusGX",
			IconURL: botAvatarURL,
		}
		embed.Timestamp = &botCreatedAt
	}

	return []discord.Embed{embed}
}

func buildComponents(request MessageRequest) []discord.ContainerComponent {
	actionRow := discord.ActionRowComponent{}

	if request.WithVote {
		actionRow.AddComponents(discord.ButtonComponent{
			Label: "Vote",
			URL:   voteLink,
			Style: discord.ButtonStyleLink,
			Emoji: &discord.ComponentEmoji{
				ID: emojiPepeHeartID,
			},
		})
	}

	if request.WithSupportServer {
		actionRow.AddComponents(discord.ButtonComponent{
			Label: "Support Server",
			URL:   supportServerLink,
			Style: discord.ButtonStyleLink,
			Emoji: &discord.ComponentEmoji{
				ID:       devEmojiID,
				Animated: true,
			},
		})
	}

	return []discord.ContainerComponent{actionRow}
}
