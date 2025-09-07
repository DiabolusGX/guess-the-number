# Guess The Number Discord Bot

Guess The Number is a simple and fun number guessing game where you have to guess a number between given minimum and maximum values for rewards that the game organizer can set. The bot can also provide hints (only if asked by game organizer).

This bot is **open source** and available on [GitHub](https://github.com/DiabolusGX/guess-the-number/)! 🌟

**Need Help?** Join our [support server](https://discord.gg/8kdx63YsDf) for assistance and updates!

**Like the bot?** Please consider [voting for us on top.gg](https://top.gg/bot/818420448131285012/vote) to help others discover it! 🗳️

**🎉 Now with Slash Commands!** - The bot has completely migrated from message commands to slash (/) commands. Try `/help` to get started.

Good Luck and Have Fun!

## 🚀 Getting Started

Setup is quite simple. You can start with **/help setup** and mainly only 1 command to setup.

### Setup Commands

You can start with `/help setup` to see what configurations you can edit & you want to set.

- _You can remove any config anytime by setting its value to `disable`_
- _Example: `/setup manager disable` will remove bot manager role._

Setup command will let you add/edit/remove bot settings:

| Command | Example | Usage |
|---------|---------|-------|
| `/setup prefix` | `/setup prefix !` | Changes prefix to `!` |
| `/setup manager` | `/setup manager @mods` | Mods role will be able to manage bot settings |
| `/setup dm` | `/setup dm enable` | Bot will DM the winner of the game |
| `/setup win-role` | `/setup win-role @winners` | Bot will add winners role to the winner |
| `/setup req-role` | `/setup req-role @level50` | Users only with this role will be eligible to guess |
| `/setup lock-role` | `/setup lock-role @level50` | Bot will lock channel for this role after game ends |
| `/setup auto-restart` | `/setup auto-restart enable` | 🔄 Auto restart game with same limits when current game ends |
| `/setup auto-reaction-hints` | `/setup auto-reaction-hints enable` | 👀 Bot reacts with ⬆️ or ⬇️ on every guess |

## 🎮 Game Commands

### Starting a Game

To **start game**, admin or bot-manager needs to enter:
`/game start <min> <max> <channel>` - Bot will start game in mentioned channel with a random number between `min` & `max` values.

- Bot will DM the answer to the game organizer (who runs the command).
- Bot will pin the game start and game end messages, lock the channel after game ends for given role and add "🔒" emoji to channel name.
- Bot will un-pin all messages sent by bot except for game-end/winning message and remove "🔒" emoji (Sometimes it can hit rate limits).

### 🔄 Auto Restart Feature

When enabled, as soon as a game ends, a new game will automatically start with the same lower and higher number limits. No need to manually start each game!

### 👀 Auto Reaction Hints

When enabled, on every guess, the bot will react with ⬆️ or ⬇️ emoji depending on whether the game's answer is higher or lower than the guess.

### Giving Hints

To **give hint** if you feel the game is too difficult:
`/game hint <first | last | number> [channel]`

- `first` - shows 1st digit of answer
- `last` - shows last digit of answer  
- `number` - Tells if answer is smaller or greater than given number

### Ending a Game

To **end game**, admin or bot-manager needs to enter:
`/game end <channel>` - Ends the currently ongoing game in mentioned channel.

- No points will be added to anyone
- Bot will not lock the channel (like it does after someone guesses the number)
- Bot will un-pin all messages sent by bot except for game-end/winning message

## 🏆 Wins & Points

When someone wins, the bot will add points to that user and add a role (if enabled by admins).

**How many points will the bot add?**

- 1/10th of the difference between `max` and `min` number in the game.

**What are points going to do?**

- Points help you climb the leaderboard (`/leaderboard`) and show your points and number of games won in your `/user info` command. Other features like purchasable roles using points are planned.

## 📊 Game Stats

Game Stats deserve their own section! Now you can see top 10 leaderboards for the following stats:

- **Closest guesses** to actual game answer
- **Most frequently guessed numbers**
- **Members with most unique numbers** guessed (no duplicate guesses)
- **Top game winners** (number of wins)
- **Top point earners**

### Viewing Stats

Stats can be viewed:

- **Per-game**: Use `/game stats <Game ID>` to see stats for a specific game
- **Server-wide**: Use `/game stats` (without Game ID) to see combined stats of all games in your server

**Finding Game ID**: Game ID can be found in game start & end messages (which are pinned in the channel where game is running). All logs about the game (start, hint, end, etc.) will also contain the Game ID.

## 👤 User Commands

Here are the commands available for users:

| Command | Example | Usage |
|---------|---------|-------|
| `/help` | `/help [command]` | List of commands OR info on 1 command |
| `/user info` | `/user info @user` | User's games won & points info |
| `/leaderboard` | `/leaderboard` | Server's leaderboard based on points |
| `/user invite` | `/user invite` | Bot invite link and other important links |
| `/game stats` | `/game stats [Game ID]` | View game statistics (per-game or server-wide) |

- Your name in Leaderboard and UserInfo will only show up after you win at least 1 game.

## 🔧 Technical Details

This Discord bot is rewritten in Go using:

- [disgo](https://github.com/disgoorg/disgo) - Discord API library
- [MongoDB Go driver](https://github.com/mongodb/mongo-go-driver) - Database
- Domain Driven Design (DDD) principles
- YAML configuration
- Structured logging with Sentry
- Prometheus metrics

### Project Structure

- `cmd/bot/`: Main entry point
- `internal/`: Business logic, domain, repository, service, bot
- `pkg/`: External/generic dependencies (logging, sentry, metrics)
- `config.yaml`: Configuration file

### Setup for Development

1. Copy `config.yaml` and fill in your settings
2. Run `go mod tidy` to install dependencies
3. Run the bot: `go run ./cmd/bot`

---

## Build and Run

`pm2-direct` make commands are recommended.
that's it for now
