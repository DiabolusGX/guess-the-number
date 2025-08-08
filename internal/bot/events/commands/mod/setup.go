package mod

import (
	"context"
	"fmt"

	"github.com/diabolusgx/guess-the-number/internal/bot/events/commands"
	"github.com/diabolusgx/guess-the-number/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"
	"github.com/disgoorg/snowflake/v2"
)

const (
	logFmt = "**Previous:** %s\n**New:** %s\n**Updated by:** %s"
)

type SetupCommand struct {
	name                   string
	guildManagementService service.GuildManagementService
}

func NewSetupCommand(params commands.CommandParams) *SetupCommand {
	return &SetupCommand{
		name:                   "setup",
		guildManagementService: params.GuildManagementService,
	}
}

func (c *SetupCommand) Name() string {
	return c.name
}

func (c *SetupCommand) Definition() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        c.name,
		Description: "Setup the bot for your server",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "prefix",
				Description: "Set the prefix for the bot",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "prefix",
						Description: "The new prefix",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "manager",
				Description: "Set the manager role for the bot",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionRole{
						Name:        "role",
						Description: "The manager role",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "dm",
				Description: "Enable or disable DMs to the winner of the game",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionBool{
						Name:        "enabled",
						Description: "Whether to enable DMs to the winner of the game",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "win-role",
				Description: "Set the role to be awarded to the winner of the game",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionRole{
						Name:        "role",
						Description: "The role to award",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "req-role",
				Description: "Set the role required to play the game",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionRole{
						Name:        "role",
						Description: "The role required to play",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "lock-role",
				Description: "Set the role to lock the channel for after the game is finished",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionRole{
						Name:        "role",
						Description: "The role to lock the channel for after the game is finished",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "log-channel",
				Description: "Set the channel where all bot activities will be logged",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: "The channel where logs will be sent",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "auto-reaction-hints",
				Description: "Enable or disable automatic reaction hints for all games",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionBool{
						Name:        "enabled",
						Description: "Whether to enable automatic reaction hints (⬆️/⬇️) for wrong guesses",
						Required:    false,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "show",
				Description: "Show all current server configuration settings",
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "info",
				Description: "Show all current server configuration settings",
			},
		},
	}
}

func (c *SetupCommand) Handler(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	subcommand := *event.SlashCommandInteractionData().SubCommandName
	switch subcommand {
	case "prefix":
		return c.handlePrefix(ctx, event, data)
	case "manager":
		return c.handleManager(ctx, event, data)
	case "dm":
		return c.handleDM(ctx, event, data)
	case "win-role":
		return c.handleWinRole(ctx, event, data)
	case "req-role":
		return c.handleReqRole(ctx, event, data)
	case "lock-role":
		return c.handleLockRole(ctx, event, data)
	case "log-channel":
		return c.handleLogChannel(ctx, event, data)
	case "auto-reaction-hints":
		return c.handleAutoReactionHints(ctx, event, data)
	case "show", "info":
		return c.handleShow(ctx, event, data)
	}
	return nil
}

func (c *SetupCommand) handlePrefix(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	prefix := event.SlashCommandInteractionData().String("prefix")

	// Validate prefix
	if len(prefix) == 0 {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Invalid Prefix",
			EmbedDescription: "Prefix cannot be empty.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	if len(prefix) > 10 {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Invalid Prefix",
			EmbedDescription: "Prefix cannot be longer than 10 characters.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Store old value for logging
	oldPrefix := data.GuildConfig.Prefix
	if oldPrefix == "" {
		oldPrefix = "gg" // Default prefix
	}

	cfg := data.GuildConfig
	cfg.Prefix = prefix
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Update Failed",
			EmbedDescription: "Failed to update prefix. Please try again later.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Log the configuration change with before/after values
	logContent := fmt.Sprintf("**Previous:** `%s`\n**New:** `%s`\n**Updated by:** <@%s>", oldPrefix, prefix, event.User().ID.String())
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel, utils.LogTypeConfigurationChange, "🔧 Prefix Updated", logContent)

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiSuccess,
		EmbedTitle:       "Prefix Updated",
		EmbedDescription: "Please use slash commands, prefix commands are deprecated.",
		EmbedColor:       utils.SuccessEmbedColor,
		IsEphemeral:      false,
		Fields: []discord.EmbedField{
			{Name: "Previous", Value: "`" + oldPrefix + "`", Inline: omit.NewPtr(true).Value},
			{Name: "New", Value: "`" + prefix + "`", Inline: omit.NewPtr(true).Value},
		},
	})
}

func (c *SetupCommand) handleManager(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	role := event.SlashCommandInteractionData().Role("role")

	role, ok := event.Client().Caches().Role(*event.GuildID(), role.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Invalid Role",
			EmbedDescription: "Given role (" + role.Mention() + ") is not there in server.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Store old value for logging
	oldManagerRole := formatRoleSetting(event, data.GuildConfig.BotManager)

	cfg := data.GuildConfig
	cfg.BotManager = role.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Update Failed",
			EmbedDescription: "Failed to update manager role. Please try again later.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Log the configuration change with before/after values
	logContent := fmt.Sprintf(logFmt, oldManagerRole, role.Mention(), event.User().Mention())
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel, utils.LogTypeConfigurationChange, "🔧 Manager Role Updated", logContent)

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiSuccess,
		EmbedTitle:       "Manager Role Updated",
		EmbedDescription: "> *Make sure to assign the role to the users who are allowed to manage the bot.*",
		EmbedColor:       utils.SuccessEmbedColor,
		IsEphemeral:      true,
		Fields: []discord.EmbedField{
			{Name: "Previous", Value: oldManagerRole, Inline: omit.NewPtr(true).Value},
			{Name: "New", Value: role.Mention(), Inline: omit.NewPtr(true).Value},
		},
	})
}

func (c *SetupCommand) handleDM(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	enabled := event.SlashCommandInteractionData().Bool("enabled")

	// Store old value for logging
	oldStatus := formatBooleanSetting(data.GuildConfig.DM)
	newStatus := formatBooleanSetting(enabled)

	cfg := data.GuildConfig
	cfg.DM = enabled
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Update Failed",
			EmbedDescription: "Failed to update DM settings. Please try again later.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Log the configuration change with before/after values
	logContent := fmt.Sprintf(logFmt, oldStatus, newStatus, event.User().Mention())
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel, utils.LogTypeConfigurationChange, "🔧 DM Settings Updated", logContent)

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiSuccess,
		EmbedTitle:       "DM Settings Updated",
		EmbedDescription: "DMs to winners will be enabled.",
		EmbedColor:       utils.SuccessEmbedColor,
		IsEphemeral:      false,
		Fields: []discord.EmbedField{
			{Name: "Previous", Value: oldStatus, Inline: omit.NewPtr(true).Value},
			{Name: "New", Value: newStatus, Inline: omit.NewPtr(true).Value},
		},
	})
}

func (c *SetupCommand) handleWinRole(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	role := event.SlashCommandInteractionData().Role("role")

	// Validate role exists in server
	_, ok := event.Client().Caches().Role(*event.GuildID(), role.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Invalid Role",
			EmbedDescription: "Given role (" + role.Mention() + ") is not there in server.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Store old value for logging
	oldWinRole := formatRoleSetting(event, data.GuildConfig.WinRole)

	cfg := data.GuildConfig
	cfg.WinRole = role.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Update Failed",
			EmbedDescription: "Failed to update win role. Please try again later.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Log the configuration change with before/after values
	logContent := fmt.Sprintf(logFmt, oldWinRole, role.Mention(), event.User().Mention())
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel, utils.LogTypeConfigurationChange, "🔧 Win Role Updated", logContent)

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiSuccess,
		EmbedTitle:       "Win Role Updated",
		EmbedDescription: "This role will be assigned to the winner of the game.",
		EmbedColor:       utils.SuccessEmbedColor,
		IsEphemeral:      false,
		Fields: []discord.EmbedField{
			{Name: "Previous", Value: oldWinRole, Inline: omit.NewPtr(true).Value},
			{Name: "New", Value: role.Mention(), Inline: omit.NewPtr(true).Value},
		},
	})
}

func (c *SetupCommand) handleReqRole(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	role := event.SlashCommandInteractionData().Role("role")

	// Validate role exists in server
	_, ok := event.Client().Caches().Role(*event.GuildID(), role.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Invalid Role",
			EmbedDescription: "Given role (" + role.Mention() + ") is not there in server.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Store old value for logging
	oldReqRole := formatRoleSetting(event, data.GuildConfig.ReqRole)

	cfg := data.GuildConfig
	cfg.ReqRole = role.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Update Failed",
			EmbedDescription: "Failed to update required role. Please try again later.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Log the configuration change with before/after values
	logContent := fmt.Sprintf(logFmt, oldReqRole, role.Mention(), event.User().Mention())
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel, utils.LogTypeConfigurationChange, "🔧 Required Role Updated", logContent)

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiSuccess,
		EmbedTitle:       "Required Role Updated",
		EmbedDescription: "> *Make sure to assign the role to the users who are allowed to play the game.*",
		EmbedColor:       utils.SuccessEmbedColor,
		IsEphemeral:      false,
		Fields: []discord.EmbedField{
			{Name: "Previous", Value: oldReqRole, Inline: omit.NewPtr(true).Value},
			{Name: "New", Value: role.Mention(), Inline: omit.NewPtr(true).Value},
		},
	})
}

func (c *SetupCommand) handleLockRole(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	role := event.SlashCommandInteractionData().Role("role")

	// Validate role exists in server
	_, ok := event.Client().Caches().Role(*event.GuildID(), role.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Invalid Role",
			EmbedDescription: "Given role (" + role.Mention() + ") is not there in server.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Store old value for logging
	oldLockRole := formatRoleSetting(event, data.GuildConfig.LockRole)

	cfg := data.GuildConfig
	cfg.LockRole = role.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Update Failed",
			EmbedDescription: "Failed to update lock role. Please try again later.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Log the configuration change with before/after values
	logContent := fmt.Sprintf(logFmt, oldLockRole, role.Mention(), event.User().Mention())
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel, utils.LogTypeConfigurationChange, "🔧 Lock Role Updated", logContent)

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiSuccess,
		EmbedTitle:       "Lock Role Updated",
		EmbedDescription: "Game channel will be locked to this role.",
		EmbedColor:       utils.SuccessEmbedColor,
		IsEphemeral:      false,
		Fields: []discord.EmbedField{
			{Name: "Previous", Value: oldLockRole, Inline: omit.NewPtr(true).Value},
			{Name: "New", Value: role.Mention(), Inline: omit.NewPtr(true).Value},
		},
	})
}

func (c *SetupCommand) handleLogChannel(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	channel := event.SlashCommandInteractionData().Channel("channel")

	// Validate channel exists in server and bot can access it
	guildChannel, ok := event.Client().Caches().Channel(channel.ID)
	if !ok {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Invalid Channel",
			EmbedDescription: "Given channel (<#" + channel.ID.String() + ">) is not accessible.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Check if channel is a text channel
	if guildChannel.Type() != discord.ChannelTypeGuildText {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Invalid Channel",
			EmbedDescription: "Log channel must be a text channel.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Store old value for logging
	oldLogChannel := formatChannelSetting(data.GuildConfig.LogChannel)

	cfg := data.GuildConfig
	cfg.LogChannel = channel.ID.String()
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		// Note: Can't log this failure to the log channel since the update failed
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Update Failed",
			EmbedDescription: "Failed to update log channel. Please try again later.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Log the configuration change to the new log channel with before/after values
	logContent := fmt.Sprintf(logFmt, oldLogChannel, guildChannel.Mention(), event.User().Mention())
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel, utils.LogTypeConfigurationChange, "🔧 Log Channel Updated", logContent)

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiSuccess,
		EmbedTitle:       "Log Channel Updated",
		EmbedDescription: "All bot activities will now be logged to this channel.",
		EmbedColor:       utils.SuccessEmbedColor,
		IsEphemeral:      false,
		Fields: []discord.EmbedField{
			{Name: "Previous", Value: oldLogChannel, Inline: omit.NewPtr(true).Value},
			{Name: "New", Value: "<#" + channel.ID.String() + ">", Inline: omit.NewPtr(true).Value},
		},
	})
}

func (c *SetupCommand) handleAutoReactionHints(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	enabled := event.SlashCommandInteractionData().Bool("enabled")

	// Store old value for logging
	oldStatus := formatBooleanSetting(data.GuildConfig.AutoReactionHints)
	newStatus := formatBooleanSetting(enabled)

	cfg := data.GuildConfig
	cfg.AutoReactionHints = enabled
	if err := c.guildManagementService.ReplaceGuildConfig(ctx, cfg); err != nil {
		return utils.EventReply(event, utils.MessageRequest{
			UseEmbed:         true,
			Emoji:            utils.EmojiError,
			EmbedTitle:       "Update Failed",
			EmbedDescription: "Failed to update auto reaction hints settings. Please try again later.",
			EmbedColor:       utils.FailureEmbedColor,
			IsEphemeral:      true,
		})
	}

	// Log the configuration change with before/after values
	logContent := fmt.Sprintf(logFmt, oldStatus, newStatus, event.User().Mention())
	utils.LogToChannel(event.Client().Rest(), cfg.LogChannel, utils.LogTypeConfigurationChange, "🔧 Auto Reaction Hints Updated", logContent)

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:         true,
		Emoji:            utils.EmojiSuccess,
		EmbedTitle:       "Auto Reaction Hints Updated",
		EmbedDescription: "> *This applies to all new games unless overridden with the /start command*",
		EmbedColor:       utils.SuccessEmbedColor,
		IsEphemeral:      false,
		Fields: []discord.EmbedField{
			{Name: "Previous", Value: oldStatus, Inline: omit.NewPtr(true).Value},
			{Name: "New", Value: newStatus, Inline: omit.NewPtr(true).Value},
		},
	})
}

func (c *SetupCommand) handleShow(ctx context.Context, event *events.ApplicationCommandInteractionCreate, data *commands.Data) error {
	cfg := data.GuildConfig

	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "gg" // Default prefix
	}

	// Prepare fields for each group
	basicSettings := ""
	// basicSettings += fmt.Sprintf("• Prefix: `%s`\n", prefix)
	// basicSettings += fmt.Sprintf("• Premium: %s\n", formatBooleanSetting(cfg.Premium))
	basicSettings += fmt.Sprintf("• DMs to Winners: %s\n", formatBooleanSetting(cfg.DM))
	basicSettings += fmt.Sprintf("• Auto Reaction Hints: %s", formatBooleanSetting(cfg.AutoReactionHints))

	roleSettings := ""
	roleSettings += fmt.Sprintf("• Manager Role: %s\n", formatRoleSetting(event, cfg.BotManager))
	roleSettings += fmt.Sprintf("• Win Role: %s\n", formatRoleSetting(event, cfg.WinRole))
	roleSettings += fmt.Sprintf("• Required Role: %s\n", formatRoleSetting(event, cfg.ReqRole))
	roleSettings += fmt.Sprintf("• Lock Role: %s", formatRoleSetting(event, cfg.LockRole))

	channelSettings := fmt.Sprintf("• Log Channel: %s", formatChannelSetting(cfg.LogChannel))

	return utils.EventReply(event, utils.MessageRequest{
		UseEmbed:          true,
		Emoji:             utils.EmojiInfo,
		EmbedTitle:        "🔧 Server Configuration",
		EmbedDescription:  "Here are the current server configuration settings:",
		EmbedColor:        utils.InfoEmbedColor,
		WithSupportServer: true,
		WithVote:          true,
		Fields: []discord.EmbedField{
			{Name: "📝 Basic Settings", Value: basicSettings, Inline: omit.NewPtr(false).Value},
			{Name: "👥 Role Settings", Value: roleSettings, Inline: omit.NewPtr(false).Value},
			{Name: "📺 Channel Settings", Value: channelSettings, Inline: omit.NewPtr(false).Value},
		},
	})
}

func formatBooleanSetting(value bool) string {
	if value {
		return "**Enabled**"
	}
	return "**Disabled**"
}

func formatRoleSetting(event *events.ApplicationCommandInteractionCreate, roleID string) string {
	if roleID == "" {
		return "❌ Not configured"
	}
	if role, ok := event.Client().Caches().Role(*event.GuildID(), snowflake.MustParse(roleID)); ok {
		return role.Mention()
	}
	return fmt.Sprintf("⚠️ <@&%s> (deleted)", roleID)
}

func formatChannelSetting(channelID string) string {
	if channelID == "" {
		return "❌ Not configured"
	}
	return fmt.Sprintf("<#%s>", channelID)
}
