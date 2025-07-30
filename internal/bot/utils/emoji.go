package utils

type Emoji int

const (
	EmojiUnspecified Emoji = iota
	EmojiSuccess
	EmojiError
)

var emojiMap = map[Emoji]string{
	EmojiSuccess: "<:greentick:768464483009691648>",
	EmojiError:   "<:redtick:768464519638024233>",
}

func (e Emoji) String() string {
	return emojiMap[e]
}
