const guildConfigModel = require("../../database/models/guildConfig");
const guildDataModel = require("../../database/models/guildData");
const GameDocument = require("../../database/documents/game");
const find = require("../../utils/find");
const { Permissions } = require("discord.js");

module.exports = {

    name: "start",
    aliases: ["go"],
    description: `Start the event with minimum and maximum number in a specific channel.`,
    category: "mod",
    guildOnly: true,
    usage: "<min> <max> [channel]",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);
        const dataDoc = await guildDataModel.findOne({ id: message.guild.id });
        const guildConfig = await guildConfigModel.findOne({ id: message.guild.id });

        if (!args[0] || !args[1] || isNaN(args[0]) || isNaN(args[1]) || parseInt(args[0]) > parseInt(args[1]))
            return message.channel.send({
                content: `${client.myEmojis[1]} | **Minimum and Maximum value missing!**\n` +
                    `Min & Max value are important to start Guess The Number Game like : \n` +
                    `- \`${dbPrefix} start 69 420 #channel\` will start game in mentioned channel with random number b/w 69 and 420.\n`
            });

        const min = parseInt(args[0]);
        const max = parseInt(args[1]);
        if (max >= Number.MAX_VALUE) return message.channel.send({
            content: `${client.myEmojis[1]} | **Maximum value is too big!**`
        });
        if (min < 0) return message.channel.send({
            content: `${client.myEmojis[1]} | **Minimum value is too small!**`
        });

        const points = Math.floor((max - min) / 10);
        let targetChannel;
        if (args[2] || message.mentions.channels.first()) {
            targetChannel = await find.getChannel(message, args[2]);
        }
        if (!targetChannel) {
            return message.channel.send({
                content: `${client.myEmojis[1]} | **Please mention a channel where you want to start the game!**\n`
            });
        }
        if (client.games.has(targetChannel.id)) return message.channel.send({
            content: `${client.myEmojis[1]} | There is already a game going on in ${targetChannel}`
        });

        const randomAnswer = Math.floor(Math.random() * (max - min) + min);

        const totalGames = dataDoc.totalGames;
        const games = dataDoc.runningGames;

        // crete new game event
        const game = await GameDocument.createGame(message.guild.id, targetChannel.id, message.author?.id ?? 0, points, randomAnswer);

        // update total games and set running game id
        games.set(targetChannel.id, { answer: randomAnswer, points: points, gameID: game._id });
        await guildDataModel.updateOne({ id: message.guild.id }, {
            $set: { runningGames: games, totalGames: totalGames + 1 }
        }, (err) => console.error);

        // set cache
        client.games.set(
            targetChannel.id,
            {
                guesses: 0,
                gameID: game._id,
                answer: randomAnswer,
            }
        );

        if (guildConfig.lockRole) {
            const role = await find.getRole(message, guildConfig.lockRole);
            targetChannel.permissionOverwrites.edit(
                role,
                { SEND_MESSAGES: true },
                { reason: "Game started, unlocked for lock role", type: 0 }
            )
                .then(channel => {
                    if (channel.name.endsWith("🔒")) channel.edit({ name: channel.name.slice(0, -1) });
                })
                .catch(console.error);
        }
        let reward = "";
        if (guildConfig.winRole) {
            const awardRole = await find.getRole(message, guildConfig.winRole);
            if (!awardRole) return message.channel.send({
                content: `${client.myEmojis[1]} | **Error while trying to find reward role.**\n` +
                    `Please remove or edit reward role by : \`${dbPrefix} setup win-role @role\``
            });
            reward = `and **${awardRole.name}** role`;
        }

        return targetChannel.send({
            content: `${client.myEmojis[0]} | **Started Game**\nMin number is \`${min}\` & Max number is \`${max}\`\n` +
                `You'll get **${points} Points** ${reward} for guessing the correct number.\n` +
                `Please vote the bot! It helps a lot <https://top.gg/bot/${client.user.id}/vote>`
        })
            .then(msg => {
                if (msg.pinnable) {
                    msg.pin({ reason: "Game Start" }).catch(console.error);
                }
                message.channel.send({
                    content: `${client.myEmojis[0]} | **Started Game** in ${targetChannel}`
                });
                return message.author.send({
                    content: `${client.myEmojis[0]} | Started game with random answer ||${randomAnswer}|| in ${targetChannel}`
                })
                    .catch(err => message.channel.send({
                        content: `Could not send game's answer to ${message.author} \n> Please Make sure your DMs are open.`
                    }));
            }).catch(console.error);
    }
}