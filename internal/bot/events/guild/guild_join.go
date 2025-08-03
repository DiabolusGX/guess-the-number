package guild

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/lib"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type GuildJoinListener struct {
	Client          bot.Client
	Logger          *logger.Logger
	GuildManagement service.GuildManagementService
}

func NewGuildJoinListener(params events.EventListenerParams) *GuildJoinListener {
	return &GuildJoinListener{
		Client:          params.Client,
		Logger:          params.Logger,
		GuildManagement: params.GuildManagementService,
	}
}

func (h *GuildJoinListener) EventName() events.EventListenerName {
	return events.GuildJoin
}

func (h *GuildJoinListener) OnEvent(ctx context.Context, e bot.Event) {
	event, ok := e.(*disgoEvents.GuildJoin)
	if !ok {
		return
	}

	ctx = context.WithValue(ctx, lib.CtxShardID, event.ShardID())
	ctx = context.WithValue(ctx, lib.CtxGuildID, event.Guild.ID.String())
	ctx = context.WithValue(ctx, lib.CtxPriority, lib.PriorityHigh)

	h.Logger.FromContext(ctx).Info("Guild joined")

	err := h.GuildManagement.CreateGuild(ctx, event.Guild.ID.String())
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("Failed to create guild", "error", err)
	}

	// fetch guild owner
	owner, err := h.Client.Rest().GetUser(event.Guild.OwnerID)
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("Failed to fetch guild owner", "error", err)
		owner = &discord.User{
			ID:       event.Guild.OwnerID,
			Username: event.Guild.OwnerID.String(),
		}
	}

	var description strings.Builder

	// guild info
	description.WriteString(fmt.Sprintf("Joined Guild: %s (`%s`)\n", event.Guild.Name, event.Guild.ID.String()))
	description.WriteString(fmt.Sprintf("Owner: %s (`%s`)\n", owner.Username, owner.ID.String()))
	description.WriteString(fmt.Sprintf("Member Count: %d\n\n", event.Guild.MemberCount))

	// shard info
	description.WriteString(fmt.Sprintf("Shard ID: %d\n", event.ShardID()))
	description.WriteString(fmt.Sprintf("Total Guilds on shard: %d\n", event.Client().Caches().GuildsLen()))
	description.WriteString(fmt.Sprintf("Total Members on shard: %d\n", event.Client().Caches().MembersAllLen()))

	guildInfoEmbed := discord.NewEmbedBuilder().
		SetColor(utils.SuccessEmbedColor).
		SetAuthor(owner.Username, "", owner.EffectiveAvatarURL()).
		SetDescription(description.String()).
		SetThumbnail(*event.Guild.IconURL()).
		SetFooterText("Guild created at").
		SetTimestamp(event.Guild.CreatedAt()).
		Build()

	messageRequest := discord.NewMessageCreateBuilder().
		SetEmbeds(guildInfoEmbed).
		Build()

	_, err = h.Client.Rest().CreateMessage(utils.AdminChannelID, messageRequest)
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("Failed to send guild info message", "error", err)
	}
}
