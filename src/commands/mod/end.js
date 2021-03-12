const guildConfigModel = require("../../database/models/guildConfig");
const guildDataModel = require("../../database/models/guildData");
const find = require("../../utils/find");

module.exports = {

    name: "end",
    aliases: ["stop"],
    description: `Ends the current going on game in specified channel.`,
    category: "mod",
    guildOnly: true,
    usage: "<channel>",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);
        const dataDoc = await guildDataModel.findOne({ id: message.guild.id });

        let targetChannel;
        if (args[0] || message.mentions.channels.first()) {
            targetChannel = await find.getChannel(message, args[2]);
        }
        else return message.channel.send(`${client.myEmojis[1]} | **Please mention a channel where you want to END the game!**\n`);
        if (!client.games.has(targetChannel.id)) return message.channel.send(`${client.myEmojis[1]} | There is NO game going on in ${targetChannel}`);

        const games = dataDoc.runningGames;
        games.delete(message.channel.id);
        client.games.delete(targetChannel.id);
        await guildDataModel.updateOne({ id: message.guild.id }, { $set: { runningGames: games } }, (err) => console.error);

        await message.channel.messages.fetchPinned()
            .then(messages => {
                messages.forEach(m => {
                    if(m.author.id === "818420448131285012" && !m.content.startsWith(`${client.myEmojis[0]} | **Congratulations**`))
                    m.unpin({ reason: "game ended" }).catch(console.error);
                })
            })
            .catch(console.error);

        return message.reply(`${client.myEmojis[0]} | **Ended Game** in ${targetChannel}`, { disableMentions: "all" });
    }
}