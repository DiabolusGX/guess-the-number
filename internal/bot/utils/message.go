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

	Emoji      Emoji
	Content    string
	Components []discord.ContainerComponent

	UseEmbed         bool
	EmbedTitle       string
	EmbedDescription string
	EmbedColor       int
	Thumbnail        string

	Fields    []discord.EmbedField
	Timestamp *string // ISO8601 or RFC3339, optional
	Footer    *discord.EmbedFooter

	WithVote          bool
	WithSupportServer bool
}

func EventReply(event *events.ApplicationCommandInteractionCreate, request MessageRequest) error {
	if request.ChannelID == 0 {
		request.ChannelID = event.Channel().ID()
	}

	// send a non-ephemeral response by deleting the ephemeral deferred response and sending a new message to the channel
	// if !request.IsEphemeral {
	// 	err := event.Client().Rest().DeleteInteractionResponse(event.ApplicationID(), event.Token())
	// 	if err != nil {
	// 		return err
	// 	}
	// 	_, err = SendMessage(event.Client().Rest(), request)
	// 	return err
	// }

	var messageUpdateRequest discord.MessageUpdate

	components := make([]discord.ContainerComponent, 0, len(request.Components))
	if len(request.Components) > 0 {
		components = append(components, request.Components...)
	}
	if derivedComponents := buildComponents(request); derivedComponents != nil {
		components = append(components, derivedComponents...)
	}

	if request.UseEmbed {
		// Create embed
		embeds := buildEmbeds(request)
		messageUpdateRequest.Embeds = &embeds
	} else {
		// Original content-based logic
		content := request.Content
		if request.Emoji != EmojiUnspecified {
			content = fmt.Sprintf(emojiContentFormat, request.Emoji.String(), request.Content)
		}
		messageUpdateRequest.Content = &content
	}

	messageUpdateRequest.Components = &components

	_, err := event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), messageUpdateRequest)
	return err
}

func SendMessage(rest rest.Rest, request MessageRequest) (*discord.Message, error) {
	var messageCreateRequest discord.MessageCreate

	components := make([]discord.ContainerComponent, 0, len(request.Components))
	if len(request.Components) > 0 {
		components = append(components, request.Components...)
	}
	if derivedComponents := buildComponents(request); derivedComponents != nil {
		components = append(components, derivedComponents...)
	}

	if request.UseEmbed {
		// Create embed
		embeds := buildEmbeds(request)
		messageCreateRequest.Embeds = embeds
	} else {
		// Original content-based logic
		content := request.Content
		if request.Emoji != EmojiUnspecified {
			content = fmt.Sprintf(emojiContentFormat, request.Emoji.String(), request.Content)
		}
		messageCreateRequest.Content = content
	}

	messageCreateRequest.Components = components

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
		Color:       DiscordBlurple,
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
	embedColor := DefaultEmbedColor
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

	if len(request.Fields) > 0 {
		embed.Fields = request.Fields
	}

	// Prepare footer
	var footerText string
	var hasFooter bool

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
	buttons := make([]discord.InteractiveComponent, 0)

	if request.WithVote {
		buttons = append(buttons,
			discord.NewLinkButton("Please vote for the bot!", voteLink).WithEmoji(discord.ComponentEmoji{
				ID:       emojiPepeHeartID,
				Animated: false,
			}),
		)
	}

	if request.WithSupportServer {
		buttons = append(buttons, discord.NewLinkButton("Join support server!", supportServerLink).WithEmoji(discord.ComponentEmoji{
			ID:       emojiDevBadgeID,
			Animated: true,
		}))
	}

	if len(buttons) == 0 {
		return nil
	}

	return []discord.ContainerComponent{discord.ActionRowComponent(buttons)}
}
