package utils

import (
	"context"
	"strings"

	"github.com/diabolusgx/guess-the-number-go/pkg/logger"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

type PermissionResult struct {
	HasAllPermissions  bool
	MissingPermissions []string
}

func CheckBotPermissions(permissions *discord.Permissions, requiredPermissions ...discord.Permissions) *PermissionResult {
	result := &PermissionResult{
		HasAllPermissions:  true,
		MissingPermissions: make([]string, 0),
	}

	for _, permission := range requiredPermissions {
		if !permissions.Has(permission) {
			result.HasAllPermissions = false
			result.MissingPermissions = append(result.MissingPermissions, permission.String())
		}
	}

	return result
}

func CheckBotPermissionsInChannel(event *events.ApplicationCommandInteractionCreate, channel discord.GuildChannel, requiredPermissions ...discord.Permissions) *PermissionResult {
	botMember, ok := event.Client().Caches().SelfMember(*event.GuildID())
	if !ok {
		return &PermissionResult{
			HasAllPermissions:  false,
			MissingPermissions: []string{"Unknown guild. Not able to get bot member"},
		}
	}

	permission := event.Client().Caches().MemberPermissionsInChannel(channel, botMember)
	return CheckBotPermissions(&permission, requiredPermissions...)
}

func UnlockChannel(ctx context.Context, event *events.ApplicationCommandInteractionCreate, channel discord.GuildChannel, lockRole discord.Role, reason string) error {
	if strings.HasSuffix(channel.Name(), "🔒") {
		name := channel.Name()[:len(channel.Name())-1]
		_, err := event.Client().Rest().UpdateChannel(
			channel.ID(),
			discord.GuildTextChannelUpdate{Name: &name},
			rest.WithReason(reason),
		)
		if err != nil {
			logger.GetLoggerFromContext(ctx).Error("Failed to unlock channel", "error", err)
		}
	}

	err := event.Client().Rest().DeletePermissionOverwrite(channel.ID(), lockRole.ID, rest.WithReason(reason))
	if err != nil {
		logger.GetLoggerFromContext(ctx).Error("Failed to delete permission overwrite", "error", err)
	}

	return err
}

func LockChannel(ctx context.Context, event *events.MessageCreate, channel discord.GuildChannel, lockRole discord.Role, reason string) error {
	if !strings.HasSuffix(channel.Name(), "🔒") {
		name := channel.Name() + "🔒"
		_, err := event.Client().Rest().UpdateChannel(
			channel.ID(),
			discord.GuildTextChannelUpdate{Name: &name},
			rest.WithReason(reason),
		)
		if err != nil {
			logger.GetLoggerFromContext(ctx).Error("Failed to lock channel", "error", err)
		}
	}

	denyPermissions := discord.PermissionSendMessages

	err := event.Client().Rest().UpdatePermissionOverwrite(
		channel.ID(),
		lockRole.ID,
		discord.RolePermissionOverwriteUpdate{
			Deny: &denyPermissions,
		},
		rest.WithReason(reason),
	)
	if err != nil {
		logger.GetLoggerFromContext(ctx).Error("Failed to create permission overwrite", "error", err)
	}

	return err
}

// CheckUserModPermissions checks if a user has permission to run mod commands.
// Returns true if user is Admin OR has the Bot Manager role.
// Makes a single REST API call to fetch member data.
func CheckUserModPermissions(event *events.ApplicationCommandInteractionCreate, botManagerRoleID string) bool {
	member, err := event.Client().Rest().GetMember(*event.GuildID(), event.User().ID)
	if err != nil {
		return false
	}

	// Check if user has administrator permissions
	permissions := event.Client().Caches().MemberPermissions(*member)
	if permissions.Has(discord.PermissionAdministrator) {
		return true
	}

	// Check if user has Bot Manager role
	if botManagerRoleID != "" {
		targetRoleID := snowflake.MustParse(botManagerRoleID)
		for _, roleId := range member.RoleIDs {
			if roleId == targetRoleID {
				return true
			}
		}
	}

	return false
}
