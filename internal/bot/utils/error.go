package utils

import (
	"context"
	"fmt"

	"github.com/diabolusgx/guess-the-number/pkg/errors"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
)

func HandleError(ctx context.Context, event ReplyEvent, err error) {
	appErr, ok := err.(*errors.AppError)
	if !ok {
		appErr = errors.New(errors.ErrCodeInternalError, err.Error())
	}

	var errReply error

	switch appErr.Code {
	case errors.ErrCodeValidation, errors.ErrCodeAlreadyExists:
		errReply = EventReply(event, MessageRequest{
			ChannelID:   event.Channel().ID(),
			Content:     fmt.Sprintf("Invalid inputs!\n%s", appErr.Message),
			Emoji:       EmojiError,
			IsEphemeral: true,
		})
	case errors.ErrCodeNotFound:
		errReply = EventReply(event, MessageRequest{
			ChannelID:   event.Channel().ID(),
			Content:     fmt.Sprintf("Not found!\n%s", appErr.Message),
			Emoji:       EmojiError,
			IsEphemeral: true,
		})
	case errors.ErrCodePermissionDenied:
		errReply = EventReply(event, MessageRequest{
			ChannelID:   event.Channel().ID(),
			Content:     fmt.Sprintf("Permission denied!\n%s", appErr.Message),
			Emoji:       EmojiError,
			IsEphemeral: true,
		})
	case errors.ErrCodeInternalError:
		errReply = EventReply(event, MessageRequest{
			ChannelID:   event.Channel().ID(),
			Content:     fmt.Sprintf("Internal Error!\n%s", appErr.Message),
			Emoji:       EmojiError,
			IsEphemeral: true,
		})
	default:
		errReply = EventReply(event, MessageRequest{
			ChannelID:   event.Channel().ID(),
			Content:     fmt.Sprintf("Error!\n%s", appErr.Message),
			Emoji:       EmojiError,
			IsEphemeral: true,
		})
	}

	if errReply != nil {
		logger.GetLoggerFromContext(ctx).Errorw("Error handling event", "error", errReply)
	}
}
