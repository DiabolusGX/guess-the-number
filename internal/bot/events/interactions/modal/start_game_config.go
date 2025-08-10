package modal

import (
	"context"
	"strconv"

	"github.com/diabolusgx/guess-the-number/internal/bot/events/commons"
	"github.com/diabolusgx/guess-the-number/internal/bot/events/interactions"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	disgoEvents "github.com/disgoorg/disgo/events"
)

type StartGameConfigModal struct {
	Logger                 *logger.Logger
	GameService            service.GameService
	GuildManagementService service.GuildManagementService
}

func NewStartGameConfigModal(params interactions.InteractionHandlerParams) *StartGameConfigModal {
	return &StartGameConfigModal{
		Logger:                 params.Logger,
		GameService:            params.GameService,
		GuildManagementService: params.GuildManagementService,
	}
}

func (m *StartGameConfigModal) CustomID() string {
	return "start_game_config"
}

func (m *StartGameConfigModal) Handler(ctx context.Context, event *disgoEvents.ModalSubmitInteractionCreate, data *interactions.Data) error {
	// Parse the configuration from the modal submission
	minStr := event.Data.Text("min")
	maxStr := event.Data.Text("max")
	autoReactionsStr := event.Data.Text("auto_reactions")

	min, err := strconv.ParseInt(minStr, 10, 64)
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "Invalid minimum number: "+minStr)
	}

	max, err := strconv.ParseInt(maxStr, 10, 64)
	if err != nil {
		return errors.New(errors.ErrCodeValidation, "Invalid maximum number: "+maxStr)
	}

	autoReactions := false
	if autoReactionsStr != "" {
		autoReactions, err = strconv.ParseBool(autoReactionsStr)
		if err != nil {
			return errors.New(errors.ErrCodeValidation, "Auto reactions must be 'true' or 'false'")
		}
	}

	// Get the target channel
	targetChannel, ok := event.Client().Caches.Channel(event.Channel().ID())
	if !ok {
		return errors.New(errors.ErrCodeInternalError, "Unable to get channel information")
	}

	// Start the game using common logic
	gameRequest := &commons.GameStartRequest{
		Client:            event.Client(),
		TargetChannel:     targetChannel,
		CreatedBy:         event.User(),
		LowerBound:        min,
		UpperBound:        max,
		AutoReactionHints: autoReactions,
	}

	result, err := commons.StartGameCommon(
		ctx,
		event.Client().Rest,
		m.GameService,
		gameRequest,
		data.GuildConfig,
	)
	if err != nil {
		utils.HandleError(ctx, event, err)
		return nil
	}

	// Format the reply message
	eventReply := commons.FormatGameStartReply(result, targetChannel, data.GuildConfig)
	return utils.EventReply(event, utils.MessageRequest{
		ChannelID:   event.Channel().ID(),
		Emoji:       utils.EmojiSuccess,
		Content:     eventReply,
		IsEphemeral: true,
	})
}
