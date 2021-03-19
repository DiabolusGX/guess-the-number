const { MessageEmbed } = require("discord.js");

module.exports = {

    name: "leaderboard",
    aliases: ["lb", "top"],
    description: "Shows points leaderboard and win games of server.",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

    },
};