package misc

import (
	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/disgoorg/disgo/bot"
)

type ReadyEventListener struct {
	Client bot.Client
	Logger *logger.Logger
}
