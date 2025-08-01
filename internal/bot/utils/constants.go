package utils

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

const (
	emojiContentFormat = "%s | %s"
	discordBlurple     = 0x5865F2
	defaultEmbedColor  = 0x99AAb5

	voteLink          = "https://top.gg/bot/818420448131285012/vote"
	supportServerLink = "https://discord.gg/QQdnQutvDd"

	botAvatarURL = "https://cdn.discordapp.com/avatars/818420448131285012/3fc53c6f1e95c0937415820aac262240.png"
)

var (
	botCreatedAt = time.Date(2021, 1, 9, 12, 11, 0, 0, time.UTC)

	emojiPepeHeartID = snowflake.MustParse("853385938930630688")
	devEmojiID       = snowflake.MustParse("818007567833497611")
)
