package lib

import (
	"context"
)

// ContextKey is a type for the keys of values stored in the context
type ContextKey string

const (
	CtxClientType ContextKey = "client_type"
	CtxRequestID  ContextKey = "request_id"
	CtxUserID     ContextKey = "user_id"
	CtxGuildID    ContextKey = "guild_id"
	CtxChannelID  ContextKey = "channel_id"
)

func GetClientType(ctx context.Context) ClientType {
	if clientType, ok := ctx.Value(CtxClientType).(ClientType); ok {
		return clientType
	}
	return ClientTypeBot
}

func GetChannelID(ctx context.Context) string {
	if channelID, ok := ctx.Value(CtxChannelID).(string); ok {
		return channelID
	}
	return ""
}

func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(CtxUserID).(string); ok {
		return userID
	}
	return ""
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(CtxRequestID).(string); ok {
		return requestID
	}
	return ""
}

func GetGuildID(ctx context.Context) string {
	if guildID, ok := ctx.Value(CtxGuildID).(string); ok {
		return guildID
	}
	return ""
}
