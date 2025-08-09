package utils

import (
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

type ReplyEvent interface {
	ApplicationID() snowflake.ID
	Token() string
	Channel() discord.InteractionChannel
	User() discord.User
	Client() bot.Client
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

	WithVote            bool
	WithSupportServer   bool
	WithStartGameButton bool
	WithBotInviteButton bool
}

func EventReply(event ReplyEvent, request MessageRequest) error {
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

	messageUpdateRequest := discord.NewMessageUpdateBuilder().SetAllowedMentions(&discord.AllowedMentions{
		Parse:       []discord.AllowedMentionType{discord.AllowedMentionTypeUsers},
		RepliedUser: true,
	})

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
		messageUpdateRequest.SetEmbeds(embeds...)
	} else {
		// Original content-based logic
		content := request.Content
		if request.Emoji != EmojiUnspecified {
			content = fmt.Sprintf(emojiContentFormat, request.Emoji.String(), request.Content)
		}
		messageUpdateRequest.SetContent(content)
	}

	messageUpdateRequest.SetContainerComponents(components...)

	_, err := event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), messageUpdateRequest.Build())
	return err
}

func SendMessage(rest rest.Rest, request MessageRequest) (*discord.Message, error) {
	messageCreateRequest := discord.NewMessageCreateBuilder().SetAllowedMentions(&discord.AllowedMentions{
		Parse:       []discord.AllowedMentionType{discord.AllowedMentionTypeUsers},
		RepliedUser: true,
	})

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
		messageCreateRequest.SetEmbeds(embeds...)
	} else {
		// Original content-based logic
		content := request.Content
		if request.Emoji != EmojiUnspecified {
			content = fmt.Sprintf(emojiContentFormat, request.Emoji.String(), request.Content)
		}
		messageCreateRequest.SetContent(content)
	}

	messageCreateRequest.SetContainerComponents(components...)

	if request.IsEphemeral {
		messageCreateRequest.SetFlags(discord.MessageFlagEphemeral)
	}

	msg, err := rest.CreateMessage(request.ChannelID, messageCreateRequest.Build())
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
func LogToChannel(rest rest.Rest, logChannelID string, logType LogType, title, description string) error {
	if logChannelID == "" {
		return nil // No log channel configured, skip logging
	}

	channelID, err := snowflake.Parse(logChannelID)
	if err != nil {
		return err
	}

	embed := discord.NewEmbedBuilder().
		SetTitle(title).
		SetDescription(description).
		SetColor(InfoEmbedColor).
		SetFooterText(logType.String()).
		Build()

	_, err = rest.CreateMessage(channelID, discord.NewMessageCreateBuilder().SetEmbeds(embed).Build())
	return err
}

// buildEmbeds creates a Discord embed from a MessageRequest
func buildEmbeds(request MessageRequest) []discord.Embed {
	isError := request.Emoji == EmojiError || request.EmbedColor == FailureEmbedColor

	embedColor := DefaultEmbedColor
	if request.EmbedColor != 0 {
		embedColor = request.EmbedColor
	}

	embedBuilder := discord.NewEmbedBuilder().SetColor(embedColor)
	if !isError {
		embedBuilder.SetThumbnail(botAvatarURL)
		embedBuilder.SetFooterText("Made with ❤️ by DiabolusGX").SetFooterIcon(botAvatarURL) //.SetTimestamp(botCreatedAt)
	}

	// Set title and description
	if request.EmbedTitle != "" {
		embedBuilder.SetTitle(request.EmbedTitle)
	}

	// Handle content based on whether we have description or content
	description := request.EmbedDescription
	if description == "" && request.Content != "" {
		description = request.Content
		if request.Emoji != EmojiUnspecified {
			description = fmt.Sprintf(emojiContentFormat, request.Emoji.String(), description)
		}
	}
	embedBuilder.SetDescription(description)

	if len(request.Fields) > 0 {
		embedBuilder.SetFields(request.Fields...)
	}

	return []discord.Embed{embedBuilder.Build()}
}

func buildComponents(request MessageRequest) []discord.ContainerComponent {
	buttons := make([]discord.InteractiveComponent, 0)

	if request.WithStartGameButton {
		buttons = append(buttons, discord.NewSuccessButton("Start New Game", "start_game").WithEmoji(discord.ComponentEmoji{
			ID:       emojiGTNLogoID,
			Animated: false,
		}))
	}

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

	if request.WithBotInviteButton {
		buttons = append(buttons, discord.NewLinkButton("Invite the bot to your server!", GetBotInviteLink(0)).WithEmoji(discord.ComponentEmoji{
			ID:       emojiGTNLogoID,
			Animated: false,
		}))
	}

	if len(buttons) == 0 {
		return nil
	}

	return []discord.ContainerComponent{discord.ActionRowComponent(buttons)}
}
