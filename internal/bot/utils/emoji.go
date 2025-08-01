package utils

type Emoji int

const (
	EmojiUnspecified Emoji = iota
	EmojiSuccess
	EmojiError
	EmojiInfo
)

var emojiMap = map[Emoji]string{
	EmojiSuccess: "<:greentick:768464483009691648>",
	EmojiError:   "<:redtick:768464519638024233>",
	EmojiInfo:    "<:arrow:1400875357711630397>",
}

func (e Emoji) String() string {
	return emojiMap[e]
}
