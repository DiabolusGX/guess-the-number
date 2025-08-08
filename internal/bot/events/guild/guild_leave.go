package guild

import (
	"context"
	"fmt"
	"strings"

	"github.com/diabolusgx/guess-the-number/internal/bot/events"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type GuildLeaveListener struct {
	Logger          *logger.Logger
	GuildManagement service.GuildManagementService
}

func NewGuildLeaveListener(params events.EventListenerParams) *GuildLeaveListener {
	return &GuildLeaveListener{
		Logger:          params.Logger,
		GuildManagement: params.GuildManagementService,
	}
}

func (h *GuildLeaveListener) EventName() events.EventListenerName {
	return events.GuildLeave
}

func (h *GuildLeaveListener) Aliases() []events.EventListenerName {
	return []events.EventListenerName{}
}

func (h *GuildLeaveListener) OnEvent(ctx context.Context, e bot.Event) error {
	event, ok := e.(*disgoEvents.GuildLeave)
	if !ok {
		return nil
	}

	ctx = context.WithValue(ctx, lib.CtxShardID, event.ShardID())
	ctx = context.WithValue(ctx, lib.CtxGuildID, event.Guild.ID.String())
	ctx = context.WithValue(ctx, lib.CtxPriority, lib.PriorityHigh)

	h.Logger.FromContext(ctx).Info("Guild left")

	err := h.GuildManagement.DeleteGuild(ctx, event.Guild.ID.String())
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("Failed to delete guild", "error", err)
	}

	// fetch guild owner
	owner, err := event.Client().Rest().GetUser(event.Guild.OwnerID)
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("Failed to fetch guild owner", "error", err)
		owner = &discord.User{
			ID:       event.Guild.OwnerID,
			Username: event.Guild.OwnerID.String(),
		}
	}

	var description strings.Builder

	// guild info
	description.WriteString(fmt.Sprintf("Left Guild: %s (`%s`)\n", event.Guild.Name, event.Guild.ID.String()))
	description.WriteString(fmt.Sprintf("Owner: %s (`%s`)\n\n", owner.Username, owner.ID.String()))
	description.WriteString(fmt.Sprintf("Member Count: %d\n", event.Guild.MemberCount))

	// shard info
	description.WriteString(fmt.Sprintf("Shard ID: %d\n", event.ShardID()))
	description.WriteString(fmt.Sprintf("Total Guilds on shard: %d\n", event.Client().Caches().GuildsLen()))
	description.WriteString(fmt.Sprintf("Total Members on shard: %d\n", event.Client().Caches().MembersAllLen()))

	guildInfoEmbed := discord.NewEmbedBuilder().
		SetColor(utils.FailureEmbedColor).
		SetAuthor(owner.Username, "", owner.EffectiveAvatarURL()).
		SetDescription(description.String()).
		SetThumbnail(*event.Guild.IconURL()).
		SetFooterText("Guild created at").
		SetTimestamp(event.Guild.CreatedAt()).
		Build()

	messageRequest := discord.NewMessageCreateBuilder().
		SetEmbeds(guildInfoEmbed).
		Build()

	_, err = event.Client().Rest().CreateMessage(utils.AdminChannelID, messageRequest)
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("Failed to send guild info message", "error", err)
	}

	return nil
}
