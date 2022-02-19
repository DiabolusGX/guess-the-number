const { MessageEmbed } = require("discord.js");
const paginationEmbed = require("../../utils/paginationEmbed");
const guildDataModel = require("../../database/models/guildData");
const GameDocument = require("../../database/documents/game");

module.exports = {

    name: "leaderboard",
    aliases: ["lb", "top"],
    description: "Shows points leaderboard and win games of server.",
    usage: "[daily | all]",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

        const winners = args[0] === "daily"
            ? await getDailyLeaderboard(message.guild.id)
            : await getAllTimeLeaderboard(message.guild.id);
        if (!winners || !winners.length) {
            return message.channel.send({
                content: `${client.myEmojis[1]} | **No one won any game yet!**`
            });
        }


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
                .setAuthor({
                    name: `${message.guild.name} has ${winners.length} Winnners`,
                    iconURL: client.user.displayAvatarURL
                })
                .setDescription(`\n${userDesc[i].substring(1)}` +
                    `[Please vote the bot! It helps a lot](https://top.gg/bot/${client.user.id}/vote)`)
        }
        return paginationEmbed(message, userEmbeds, emojiList, 30000);
    },
};

async function getDailyLeaderboard(guildID) {
    const games = await GameDocument.getDailyLeaderboard(guildID);

    // count wins and total points of each winner
    const winnersData = games.reduce((acc, cur) => {
        if (!acc[cur.wonBy]) {
            acc[cur.wonBy] = {
                wins: 0,
                points: 0
            };
        }
        acc[cur.wonBy].wins += 1;
        acc[cur.wonBy].points += cur.points;
        return acc;
    }, {});

    const winners = [];
    for (const userID in winnersData) {
        const data = winnersData[userID];
        winners.push(`<@${userID}> [**${data.wins}x**] with **${data.points}** Points\n`);
    }

    return winners;
}

async function getAllTimeLeaderboard(guildID) {
    const dataDoc = await guildDataModel.findOne({ id: guildID }).catch(console.error);
    const users = dataDoc.users;

    const tempData = [], winners = [];

    users.forEach((value, key) => {
        value.user = key;
        tempData.push(value);
    });

    if (!tempData.length) {
        return false;
    }

    tempData.sort((a, b) => b.points - a.points);

    for (const data of tempData) {
        winners.push(`<@${data.user}> [**${data.wins}x**] with **${data.points}** Points\n`);
    }

    return winners;
}
