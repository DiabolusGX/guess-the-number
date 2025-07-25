package mod

import (
	"context"
	"fmt"

	"github.com/diabolusgx/guess-the-number-go/internal/bot/commands"
	"github.com/diabolusgx/guess-the-number-go/internal/bot/utils"
	"github.com/diabolusgx/guess-the-number-go/internal/service"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type SetupCommand struct {
	name                   string
	definition             discord.ApplicationCommandCreate
	guildManagementService service.GuildManagementService
}

func NewSetupCommand(params commands.CommandParams) *SetupCommand {
	return &SetupCommand{
		name: "setup",
		definition: discord.SlashCommandCreate{
			Name:        "setup",
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
					Description: "Enable or disable DMs from the bot",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionBool{
							Name:        "enabled",
							Description: "Whether to enable DMs",
							Required:    true,
						},
					},
				},
				discord.ApplicationCommandOptionSubCommand{
					Name:        "win-role",
					Description: "Set the role to be awarded to the winner",
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
					Description: "Set the role to lock the channel for after a win",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionRole{
							Name:        "role",
							Description: "The role to lock the channel for",
							Required:    true,
						},
					},
				},
			},
		},
		guildManagementService: params.GuildManagementService,
	}
}

func (c *SetupCommand) Name() string {
	return c.name
}

func (c *SetupCommand) Definition() discord.ApplicationCommandCreate {
	return c.definition
}

func (c *SetupCommand) Handler(event *events.ApplicationCommandInteractionCreate) error {
	subcommand := *event.SlashCommandInteractionData().SubCommandName
	switch subcommand {
	case "prefix":
		return c.handlePrefix(event)
	case "manager":
		return c.handleManager(event)
	case "dm":
		return c.handleDM(event)
	case "win-role":
		return c.handleWinRole(event)
	case "req-role":
		return c.handleReqRole(event)
	case "lock-role":
		return c.handleLockRole(event)
	}
	return nil
}

func (c *SetupCommand) handlePrefix(event *events.ApplicationCommandInteractionCreate) error {
	prefix := event.SlashCommandInteractionData().String("prefix")
	cfg, err := c.guildManagementService.GetGuildConfig(context.Background(), event.GuildID().String())
	if err != nil {
		return err
	}

	cfg.Prefix = prefix
	if _, err := c.guildManagementService.ReplaceGuildConfig(context.Background(), cfg); err != nil {
		return err
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Prefix Updated", fmt.Sprintf("Prefix updated to `%s`", prefix), 0x00FF00)
}

func (c *SetupCommand) handleManager(event *events.ApplicationCommandInteractionCreate) error {
	role := event.SlashCommandInteractionData().Role("role")
	cfg, err := c.guildManagementService.GetGuildConfig(context.Background(), event.GuildID().String())
	if err != nil {
		return err
	}

	cfg.BotManager = role.ID.String()
	if _, err := c.guildManagementService.ReplaceGuildConfig(context.Background(), cfg); err != nil {
		return err
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Manager Role Updated", fmt.Sprintf("Manager role updated to %s", role.Mention()), 0x00FF00)
}

func (c *SetupCommand) handleDM(event *events.ApplicationCommandInteractionCreate) error {
	enabled := event.SlashCommandInteractionData().Bool("enabled")
	cfg, err := c.guildManagementService.GetGuildConfig(context.Background(), event.GuildID().String())
	if err != nil {
		return err
	}

	cfg.DM = enabled
	if _, err := c.guildManagementService.ReplaceGuildConfig(context.Background(), cfg); err != nil {
		return err
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}
	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "DMs Updated", fmt.Sprintf("DMs have been %s", status), 0x00FF00)
}

func (c *SetupCommand) handleWinRole(event *events.ApplicationCommandInteractionCreate) error {
	role := event.SlashCommandInteractionData().Role("role")
	cfg, err := c.guildManagementService.GetGuildConfig(context.Background(), event.GuildID().String())
	if err != nil {
		return err
	}

	cfg.WinRole = role.ID.String()
	if _, err := c.guildManagementService.ReplaceGuildConfig(context.Background(), cfg); err != nil {
		return err
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Win Role Updated", fmt.Sprintf("Win role updated to %s", role.Mention()), 0x00FF00)
}

func (c *SetupCommand) handleReqRole(event *events.ApplicationCommandInteractionCreate) error {
	role := event.SlashCommandInteractionData().Role("role")
	cfg, err := c.guildManagementService.GetGuildConfig(context.Background(), event.GuildID().String())
	if err != nil {
		return err
	}

	cfg.ReqRole = role.ID.String()
	if _, err := c.guildManagementService.ReplaceGuildConfig(context.Background(), cfg); err != nil {
		return err
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Required Role Updated", fmt.Sprintf("Required role updated to %s", role.Mention()), 0x00FF00)
}

func (c *SetupCommand) handleLockRole(event *events.ApplicationCommandInteractionCreate) error {
	role := event.SlashCommandInteractionData().Role("role")
	cfg, err := c.guildManagementService.GetGuildConfig(context.Background(), event.GuildID().String())
	if err != nil {
		return err
	}

	cfg.LockRole = role.ID.String()
	if _, err := c.guildManagementService.ReplaceGuildConfig(context.Background(), cfg); err != nil {
		return err
	}

	return utils.SendEmbed(event.Client().Rest(), event.Channel().ID(), "Lock Role Updated", fmt.Sprintf("Lock role updated to %s", role.Mention()), 0x00FF00)
}
