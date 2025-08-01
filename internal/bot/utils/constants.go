package utils

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

const (
	// Links
	voteLink          = "https://top.gg/bot/818420448131285012/vote"
	supportServerLink = "https://discord.gg/QQdnQutvDd"

	// Colors
	DiscordBlurple    = 0x5865F2
	DefaultEmbedColor = 0x2B2D31
	SuccessEmbedColor = 0x57F287
	FailureEmbedColor = 0xED4245
	InfoEmbedColor    = 0x5865F2

	// Bot info
	botAvatarURL = "https://cdn.discordapp.com/avatars/818420448131285012/3fc53c6f1e95c0937415820aac262240.png"

	// Emoji
	emojiContentFormat = "%s | %s"
	emojiPepeHeartID   = snowflake.ID(853385938930630688)
	emojiDevBadgeID    = snowflake.ID(818007567833497611)
)

var (
	botCreatedAt = time.Date(2021, 1, 8, 12, 18, 51, 0, time.UTC)
)
