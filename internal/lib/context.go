package lib

import (
	"context"
	"strconv"
)

// ContextKey is a type for the keys of values stored in the context
type ContextKey string

const (
	CtxClientType ContextKey = "client_type"
	CtxShardID    ContextKey = "shard_id"
	CtxRequestID  ContextKey = "request_id"
	CtxPriority   ContextKey = "priority"
	CtxUserID     ContextKey = "user_id"
	CtxGuildID    ContextKey = "guild_id"
	CtxChannelID  ContextKey = "channel_id"
	CtxGameID     ContextKey = "game_id"
	CtxFlowID     ContextKey = "flow_id"
)

var contextKeys = []ContextKey{CtxClientType, CtxShardID, CtxRequestID, CtxPriority, CtxUserID, CtxGuildID, CtxChannelID, CtxGameID}

type Priority int

const (
	PriorityDefault Priority = iota
	PriorityHigh
	PriorityCritical
)

func (p Priority) String() string {
	switch p {
	case PriorityDefault:
		return "default"
	case PriorityHigh:
		return "high"
	case PriorityCritical:
		return "critical"
	}
	return strconv.Itoa(int(p))
}

func GetPriority(ctx context.Context) Priority {
	if priority, ok := ctx.Value(CtxPriority).(Priority); ok {
		return priority
	}
	return PriorityDefault
}

func GetClientType(ctx context.Context) ClientType {
	if clientType, ok := ctx.Value(CtxClientType).(ClientType); ok {
		return clientType
	}
	return ClientTypeBot
}

func GetShardID(ctx context.Context) int {
	if shardID, ok := ctx.Value(CtxShardID).(int); ok {
		return shardID
	}
	return 0
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

func GetGameID(ctx context.Context) string {
	if gameID, ok := ctx.Value(CtxGameID).(string); ok {
		return gameID
	}
	return ""
}

func GetFlowID(ctx context.Context) string {
	if flowID, ok := ctx.Value(CtxFlowID).(string); ok {
		return flowID
	}
	return ""
}

func CopyContextKeys(ctx context.Context) context.Context {
	newCtx := context.Background()
	for _, key := range contextKeys {
		newCtx = context.WithValue(newCtx, key, ctx.Value(key))
	}
	return newCtx
}
