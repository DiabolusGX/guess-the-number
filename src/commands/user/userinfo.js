const { MessageEmbed } = require("discord.js");

module.exports = {

    name: "userinfo",
    aliases: ["info"],
    description: "Shows user's win games, points and other basic info.",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

    },
};