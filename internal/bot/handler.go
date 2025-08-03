package bot

import (
	"context"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/events"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/config"
	"github.com/diabolusgx/guess-the-number-go/internal/lib"
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/diabolusgx/guess-the-number-go/pkg/metrics"
	"github.com/disgoorg/disgo/bot"
	"go.uber.org/fx"
)

type BotHandlerParams struct {
	fx.In

	Config  *config.Configuration
	Logger  *logger.Logger
	Metrics *metrics.Metrics

	ApplicationCommandHandler *commands.ApplicationCommandHandler
}

type BotHandler interface {
	bot.EventListener

	AddEventListener(listener events.Listener)
	SyncCommands() error
}

type handler struct {
	logger  *logger.Logger
	config  *config.Configuration
	metrics *metrics.Metrics

	applicationCommandHandler *commands.ApplicationCommandHandler
	listeners                 map[events.EventListenerName]events.Listener
}

func NewBotHandler(params BotHandlerParams) BotHandler {
	return &handler{
		logger:  params.Logger,
		config:  params.Config,
		metrics: params.Metrics,

		applicationCommandHandler: params.ApplicationCommandHandler,

		listeners: make(map[events.EventListenerName]events.Listener),
	}
}

func (h *handler) AddEventListener(listener events.Listener) {
	h.listeners[listener.EventName()] = listener
	for _, alias := range listener.Aliases() {
		h.listeners[alias] = listener
	}
}

func (h *handler) OnEvent(event bot.Event) {
	if event == nil {
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, lib.CtxRequestID, lib.NewRequestID())

	// panic recovery
	defer func() {
		if r := recover(); r != nil {
			h.logger.FromContext(ctx).Errorw("panic in event listener", "err", r)
		}
	}()

	// Track event metrics
	eventName := string(events.GetEventListenerName(event))
	success := "true"

	listener, ok := h.listeners[events.GetEventListenerName(event)]
	if !ok {
		return
	}

	err := listener.OnEvent(ctx, event)
	if err != nil {
		ctx = context.WithValue(ctx, lib.CtxPriority, lib.PriorityCritical)
		h.logger.FromContext(ctx).Errorw("error in event listener", "event", eventName, "err", err)
		success = "false"
	}

	h.metrics.Operations.WithLabelValues("event", eventName, success).Inc()
}

func (h *handler) SyncCommands() error {
	return h.applicationCommandHandler.SyncCommands()
}
