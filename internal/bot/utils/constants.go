package utils

import (
	"fmt"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

const (
	// Links
	voteLink          = "https://top.gg/bot/818420448131285012/vote"
	supportServerLink = "https://discord.gg/QQdnQutvDd"

	// Colors
	DefaultEmbedColor = 0x2B2D31
	SuccessEmbedColor = 0x57F287
	FailureEmbedColor = 0xED4245
	WarningEmbedColor = 0xFEE75C // Discord yellow
	InfoEmbedColor    = 0x5865F2 // Discord blurple

	// Bot info
	botAvatarURL = "https://cdn.discordapp.com/avatars/818420448131285012/3fc53c6f1e95c0937415820aac262240.png"

	// Emoji
	emojiContentFormat = "%s | %s"
	emojiPepeHeartID   = snowflake.ID(853385938930630688)
	emojiDevBadgeID    = snowflake.ID(818007567833497611)
	emojiGTNLogoID     = snowflake.ID(1401528921694015649)

	// Admin
	AdminChannelID = snowflake.ID(818439901044801567)
)

var (
	botCreatedAt = time.Date(2021, 1, 8, 12, 18, 51, 0, time.UTC)
)

type LogType string

const (
	LogTypeGameActivity        LogType = "Game Activity"
	LogTypeConfigurationChange LogType = "Configuration Change"
)

func (t LogType) String() string {
	return string(t)
}

func FormMessageLink(guildID, channelID, messageID string) string {
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, messageID)
}

func GetBotInviteLink(clientID snowflake.ID) string {
	if clientID == 0 {
		clientID = defaultBotClientID
	}

	return fmt.Sprintf("https://discord.com/oauth2/authorize?client_id=%s", clientID.String())
}
