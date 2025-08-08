package interactions

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"

	"github.com/diabolusgx/guess-the-number/internal/bot/events"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/internal/lib"
	"github.com/diabolusgx/guess-the-number/internal/service"
	ierr "github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"github.com/diabolusgx/guess-the-number/pkg/metrics"
)

type InteractionHandler struct {
	Config  *config.Configuration
	Logger  *logger.Logger
	Metrics *metrics.Metrics

	GameService            service.GameService
	GuildManagementService service.GuildManagementService

	ComponentInteractions map[string]ComponentInteraction
	ModalInteractions     map[string]ModalInteraction
}

func NewInteractionHandler(params events.EventListenerParams) *InteractionHandler {
	return &InteractionHandler{
		Config:  params.Config,
		Logger:  params.Logger,
		Metrics: params.Metrics,

		GameService:            params.GameService,
		GuildManagementService: params.GuildManagementService,

		ComponentInteractions: make(map[string]ComponentInteraction),
		ModalInteractions:     make(map[string]ModalInteraction),
	}
}

func (h *InteractionHandler) EventName() events.EventListenerName {
	// using alias to avoid creating a new event listener for each interaction type
	return events.EventUnspecified
}

func (h *InteractionHandler) Aliases() []events.EventListenerName {
	return []events.EventListenerName{
		events.ComponentInteraction,
		events.ModalSubmit,
	}
}

func (h *InteractionHandler) OnEvent(ctx context.Context, e bot.Event) error {
	switch event := e.(type) {
	case *disgoEvents.ComponentInteractionCreate:
		// TODO: re-visit this
		// err := event.DeferCreateMessage(false)
		// if err != nil {
		// 	h.Logger.FromContext(ctx).Errorw("error acknowledging discord interaction", "err", err)
		// }

		return h.handleComponentInteraction(ctx, event)
	case *disgoEvents.ModalSubmitInteractionCreate:
		// TODO: re-visit this
		err := event.DeferCreateMessage(false)
		if err != nil {
			h.Logger.FromContext(ctx).Errorw("error acknowledging discord interaction", "err", err)
		}

		return h.handleModalInteraction(ctx, event)
	default:
		return nil
	}
}

func (h *InteractionHandler) AddComponentInteraction(interaction ComponentInteraction) {
	h.ComponentInteractions[interaction.CustomID()] = interaction
}

func (h *InteractionHandler) AddModalInteraction(interaction ModalInteraction) {
	h.ModalInteractions[interaction.CustomID()] = interaction
}

func (h *InteractionHandler) handleComponentInteraction(ctx context.Context, event *disgoEvents.ComponentInteractionCreate) error {
	// Add context values
	ctx = context.WithValue(ctx, lib.CtxShardID, event.ShardID())
	ctx = context.WithValue(ctx, lib.CtxGuildID, event.GuildID().String())
	ctx = context.WithValue(ctx, lib.CtxChannelID, event.Channel().ID().String())
	ctx = context.WithValue(ctx, lib.CtxUserID, event.User().ID.String())

	customID := event.Data.CustomID()
	var interactionSuccess = "true"

	err := h.processComponentInteraction(ctx, event)
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("error handling component interaction", "err", err, "custom_id", customID)
		interactionSuccess = "false"
	}

	h.Metrics.Operations.WithLabelValues("component_interaction", customID, interactionSuccess).Inc()
	return nil
}

func (h *InteractionHandler) handleModalInteraction(ctx context.Context, event *disgoEvents.ModalSubmitInteractionCreate) error {
	// Add context values
	ctx = context.WithValue(ctx, lib.CtxShardID, event.ShardID())
	ctx = context.WithValue(ctx, lib.CtxGuildID, event.GuildID().String())
	ctx = context.WithValue(ctx, lib.CtxChannelID, event.Channel().ID().String())
	ctx = context.WithValue(ctx, lib.CtxUserID, event.User().ID.String())

	customID := event.Data.CustomID
	var interactionSuccess = "true"

	err := h.processModalInteraction(ctx, event)
	if err != nil {
		h.Logger.FromContext(ctx).Errorw("error handling modal interaction", "err", err, "custom_id", customID)
		interactionSuccess = "false"
	}

	h.Metrics.Operations.WithLabelValues("modal_interaction", customID, interactionSuccess).Inc()
	return nil
}

func (h *InteractionHandler) processComponentInteraction(ctx context.Context, event *disgoEvents.ComponentInteractionCreate) error {
	customID := event.Data.CustomID()

	// Check bot permissions
	appPermissions := event.AppPermissions()
	res := utils.CheckBotPermissions(appPermissions, discord.PermissionViewChannel, discord.PermissionSendMessages, discord.PermissionEmbedLinks)
	if !res.HasAllPermissions {
		h.Logger.FromContext(ctx).Infow("missing permissions", "missing_permissions", strings.Join(res.MissingPermissions, ", "))
		return ierr.New(ierr.ErrCodeMissingPermissions, "bot is missing permissions")
	}

	guildConfig, err := h.GuildManagementService.GetGuildConfig(ctx, event.GuildID().String())
	if err != nil {
		utils.HandleError(ctx, event, err)
		return err
	}

	// Check permissions for mod component interactions
	if isModComponentInteraction(customID) {
		if !utils.CheckUserModPermissionsComponent(event, guildConfig.BotManager) {
			content := "❌ **Access Denied**\nYou need to be an Administrator or have the Bot Manager role to use this action."
			_ = event.CreateMessage(discord.MessageCreate{
				Content: content,
				Flags:   discord.MessageFlagEphemeral,
			})
			h.Logger.FromContext(ctx).Infow("user lacks permissions for mod component interaction", "custom_id", customID, "user_id", event.User().ID.String())
			return ierr.New(ierr.ErrCodeMissingPermissions, "user is missing permissions")
		}
	}

	commonData := &Data{
		GuildConfig: guildConfig,
	}

	interaction, ok := h.ComponentInteractions[customID]
	if !ok {
		return ierr.NewErrorWithContext(ctx, ierr.ErrCodeNotFound, fmt.Errorf("component interaction not found: %s", customID))
	}

	if err := interaction.Handler(ctx, event, commonData); err != nil {
		utils.HandleError(ctx, event, err)
		return err
	}

	return nil
}

func (h *InteractionHandler) processModalInteraction(ctx context.Context, event *disgoEvents.ModalSubmitInteractionCreate) error {
	customID := event.Data.CustomID

	// Check bot permissions
	appPermissions := event.AppPermissions()
	res := utils.CheckBotPermissions(appPermissions, discord.PermissionViewChannel, discord.PermissionSendMessages, discord.PermissionEmbedLinks)
	if !res.HasAllPermissions {
		h.Logger.FromContext(ctx).Infow("missing permissions", "missing_permissions", strings.Join(res.MissingPermissions, ", "))
		return ierr.New(ierr.ErrCodeMissingPermissions, "bot is missing permissions")
	}

	guildConfig, err := h.GuildManagementService.GetGuildConfig(ctx, event.GuildID().String())
	if err != nil {
		utils.HandleError(ctx, event, err)
		return err
	}

	// Check permissions for mod modal interactions
	if isModModalInteraction(customID) {
		if !utils.CheckUserModPermissionsModal(event, guildConfig.BotManager) {
			content := "❌ **Access Denied**\nYou need to be an Administrator or have the Bot Manager role to use this action."
			_ = event.CreateMessage(discord.MessageCreate{
				Content: content,
				Flags:   discord.MessageFlagEphemeral,
			})
			h.Logger.FromContext(ctx).Infow("user lacks permissions for mod modal interaction", "custom_id", customID, "user_id", event.User().ID.String())
			return ierr.New(ierr.ErrCodeMissingPermissions, "user is missing permissions")
		}
	}

	commonData := &Data{
		GuildConfig: guildConfig,
	}

	interaction, ok := h.ModalInteractions[customID]
	if !ok {
		return ierr.NewErrorWithContext(ctx, ierr.ErrCodeNotFound, fmt.Errorf("modal interaction not found: %s", customID))
	}

	if err := interaction.Handler(ctx, event, commonData); err != nil {
		utils.HandleError(ctx, event, err)
		return err
	}

	return nil
}

// isModComponentInteraction checks if a component interaction customID is a mod interaction
func isModComponentInteraction(customID string) bool {
	modComponentInteractions := []string{"start_game"}
	return slices.Contains(modComponentInteractions, customID)
}

// isModModalInteraction checks if a modal interaction customID is a mod interaction
func isModModalInteraction(customID string) bool {
	modModalInteractions := []string{"start_game_config"}
	return slices.Contains(modModalInteractions, customID)
}
