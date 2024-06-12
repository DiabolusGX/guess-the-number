const guildDataModel = require("../../database/models/guildData");
const GameDocument = require("../../database/documents/game");
const find = require("../../utils/find");

module.exports = {

    name: "end",
    aliases: ["stop"],
    description: "Ends the currently ongoing game in mentioned channel",
    category: "mod",
    guildOnly: true,
    usage: "<channel>",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);
        const dataDoc = await guildDataModel.findOne({ id: message.guild.id });

        let targetChannel;
        if (args[0] || message.mentions.channels.first()) {
            targetChannel = await find.getChannel(message, args[0]);
        }
        else return message.channel.send({
            content: `${client.myEmojis[1]} | **Please mention a channel where you want to END the game!**\n`
        });
        if (!client.games.has(targetChannel.id)) return message.channel.send({
            content: `${client.myEmojis[1]} | There is NO game going on in ${targetChannel}`
        });

        // clear cache game
        const gameData = client.games.get(message.channel.id);
        const { gameID, guesses } = gameData;
        client.games.delete(targetChannel.id);

        // update DB
        const games = dataDoc.runningGames;
        games.delete(message.channel.id);
        await guildDataModel.updateOne({ id: message.guild.id }, { $set: { runningGames: games } }, (err) => console.error);
        await GameDocument.disableGame(gameID, guesses);

        await message.channel.messages.fetchPinned()
            .then(messages => {
                messages.forEach(m => {
                    if (m.author.id === client.user.id && !m.content.startsWith(`${client.myEmojis[0]} | **Congratulations**`))
                        m.unpin({ reason: "game ended" }).catch(console.error);
                })
            })
            .catch(console.error);

        return message.reply(`${client.myEmojis[0]} | **Ended Game** in ${targetChannel}`, { disableMentions: "all" });
    }
}
