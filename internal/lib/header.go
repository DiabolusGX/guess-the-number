package lib

const (
	HeaderRequestID = "X-Request-ID"
	HeaderUserID    = "X-User-ID"
	HeaderGuildID   = "X-Guild-ID"
	HeaderChannelID = "X-Channel-ID"
)

// ClientType denotes the type of the client
type ClientType string

const (
	// ClientTypeBot denotes a bot client
	ClientTypeBot ClientType = "bot"
	// ClientTypeWeb denotes a web client (for future web interface)
	ClientTypeWeb ClientType = "web"
)

func (c ClientType) String() string {
	return string(c)
}
