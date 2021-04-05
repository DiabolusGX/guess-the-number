const { MessageEmbed } = require("discord.js");
const paginationEmbed = require("../../utils/paginationEmbed");
const guildDataModel = require("../../database/models/guildData");

module.exports = {

    name: "leaderboard",
    aliases: ["lb", "top"],
    description: "Shows points leaderboard and win games of server.",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

        const dataDoc = await guildDataModel.findOne({ id: message.guild.id }).catch(console.error);
        const users = dataDoc.users;
        const tempData = [], winners = [];
        users.forEach((value, key) => {
            value.user = key;
            tempData.push(value);
        });

        if(!tempData.length) return message.channel.send(`${client.myEmojis[1]} | **No one won any game yet!**`)

        tempData.sort((a, b) => b.points - a.points);
        tempData.forEach(data => {
            winners.push(`<@${data.user}> [**${data.wins}x**] with **${data.points}** Points\n`);
        });

        let noOfEmbeds, userDesc = [];
        if ((winners.length % 10) > 0) noOfEmbeds = parseInt((winners.length / 10) + 1);
        else noOfEmbeds = parseInt((winners.length / 10));

        for (let i = 0; i < noOfEmbeds; i++) {
            userDesc[i] = "\n";
            for (let j = 0; j < 10; j++) {
                if ((i * 10 + j) === winners.length) break;
                userDesc[i] += `\`${i * 10 + j + 1}.\` ${winners[i * 10 + j]}\n`;
            }
        }

        const userEmbeds = [], emojiList = ["⏪", "⏩"];
        for (let i = 0; i < noOfEmbeds; i++) {
            userEmbeds[i] = new MessageEmbed()
                .setColor(client.colors[0])
                .setAuthor(`${message.guild.name} has ${winners.length} Winnners`, client.user.displayAvatarURL)
                .setDescription(`\n${userDesc[i].substring(1)}`+
                    `[Please vote the bot! It helps a lot](https://top.gg/bot/818420448131285012/vote)`)
        }
        return paginationEmbed(message, userEmbeds, emojiList, 30000);
    },
};