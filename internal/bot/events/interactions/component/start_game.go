package component

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/interactions"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/diabolusgx/guess-the-number-go/pkg/errors"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/disgoorg/disgo/discord"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type StartGameInteraction struct {
	Logger      *logger.Logger
	GameService service.GameService
}

func NewStartGameInteraction(params interactions.InteractionHandlerParams) *StartGameInteraction {
	return &StartGameInteraction{
		Logger:      params.Logger,
		GameService: params.GameService,
	}
}

func (i *StartGameInteraction) CustomID() string {
	return "start_game"
}

func (i *StartGameInteraction) Handler(ctx context.Context, event *disgoEvents.ComponentInteractionCreate, data *interactions.Data) error {
	// Build modal with game configuration (using default values)
	modal := discord.NewModalCreateBuilder().
		SetCustomID("start_game_config").
		SetTitle("Start New Game Configuration").
		AddActionRow(
			discord.NewShortTextInput("min", "Minimum Number").
				WithRequired(true).
				WithPlaceholder("e.g., 1").
				WithMinLength(1).
				WithMaxLength(10),
		).
		AddActionRow(
			discord.NewShortTextInput("max", "Maximum Number").
				WithRequired(true).
				WithPlaceholder("e.g., 100").
				WithMinLength(1).
				WithMaxLength(10),
		).
		AddActionRow(
			discord.NewShortTextInput("auto_reactions", "Auto Reaction Hints (true/false)").
				WithRequired(false).
				WithPlaceholder("true").
				WithValue("false").
				WithMinLength(4).
				WithMaxLength(5),
		).
		Build()

	err := event.Modal(modal)
	if err != nil {
		i.Logger.FromContext(ctx).Error("failed to create modal", "error", err.Error())
		return errors.WithError(err).Mark(errors.ErrCodeInternalError)
	}

	return nil
}
