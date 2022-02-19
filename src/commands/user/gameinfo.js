const { MessageEmbed } = require("discord.js");
const guildDataModel = require("../../database/models/guildData");
const find = require("../../utils/find");

module.exports = {

    name: "gameinfo",
    aliases: ["game", "gi"],
    description: "Show running game info in spefific mentioned channel.",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

        let targetChannel;
        if (args[0] || message.mentions.channels.first()) {
            targetChannel = await find.getChannel(message, args[0]);
        }
        else targetChannel = message.channel;

        const guildData = await guildDataModel.findOne({ id: message.guild.id }).catch(console.error);
        const totalGames = guildData?.totalGames;
        let infoDesc = "";
        if (totalGames) {
            infoDesc = `Total **${totalGames}** games played in ${message.guild.name} 🎉\n\n`;
        }
        else infoDesc = `NO games are running in ${message.guild.name}!\n\n` +
            `Ask any mod to use \`${dbPrefix} start\` to start the game for you 🎉`;

        const gameInfo = client.games.get(targetChannel.id);
        if (gameInfo) {
            const games = guildData?.runningGames;
            const { guesses } = gameInfo;
            const { points } = games?.get(targetChannel.id);

            infoDesc += `Game in ${targetChannel} for **${points}** Points has **${guesses}** guesses so far ⚖️`;
        }
        else {
            infoDesc += `But currently NO game is running in ${targetChannel} ⚖️\n` +
                `Ask any mod to use \`${dbPrefix} start\` to start the game for you 🎉`;
        }

        const gameInfoEmbed = new MessageEmbed()
            .setColor(client.colors[0])
            .setAuthor({
                name: message.member.nickname || message.author.username,
                iconURL: message.author.displayAvatarURL({ format: "png", dynamic: true })
            })
            .setDescription(`${infoDesc}`)

        return message.channel.send({ embeds: [gameInfoEmbed] });
    },
};