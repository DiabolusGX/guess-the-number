# Guess The Number

Hello there!
Guess The Number will be a simple and fune number game where you have to guess number b/w given minimum and maximum for some reward that game organizer can set.
Bot can also provide hints (only if asked by game organizer) but that'll lead to less game win points.
Good Luck and Have Fun!


# Getting Started

Setup is quite simple. You can start with **bb help setup** and mainly only 1 command to setup.

## Setup

You can start with `gg help setup` to see what configurations you can edit & you want to set.
- _You can remove any config anytime by setting it's value to  `disable`_
_ex :  `bb setup manager disable`  will remove bot manager role._

Setup command will let you add/edit/remove bot settings like :

| Command   |      Example |  Usage |
|----------|:-------------|:-------|
| prefix |  `gg setup prefix !` | It'll change prefix to `!` |
| manager|  `gg setup manager @mods` | Mods role will be able to manage bot settings |
| DM |  `gg setup dm enable` | Bot will DM to the winner of the game |
| win-role|  `gg setup win-role @winners` | Bot will add winners role to the winner |
| req-role|  `gg setup req-role @level50` | Users only with this role will be eligible to guess the number |
| lock-role|  `gg setup lock-role @level50` | Bot will lock channel for this role after game is over |

## Start Game

To **start game** admin or bot-manager need to enter just 1 command :
`gg start <min> <max> [channel]` - Bot will start game in mentioned channel (current channel, if not mentioned) with a random number between `min` & `max` value provided.
- Bot will DM the answer to the game organizer (who runs the command).
- Bot will pin the game start and game end message, lock the channel after game ends for given role and add "🔒" emoji to channel name.
- Bot will un-pin all messages sent by bot except for game-end/winning message and remove "🔒" emoji (Sometimes it can hit rate limits).

To **give hint** if you feel game is too difficult you can use :
`gg hint <first | last | number> [channel]`
- `first` - shows 1st digit of answer.
- `last` - shows last digit of answer.
- `number` - Tells if answer is smaller or greater then given number.

## Wins

When someone wins bot will add points to thet users and add a role (If enabled by admins)
- How many points bot will add?
- - 1/10 th of the difference between `max` and `min` number in game.

- What are points gonna do?
- - Nothing as of now, but there is gonna be a leaderboard and some purchasable items or roles with points.

