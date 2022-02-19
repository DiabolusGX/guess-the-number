const { MessageEmbed } = require("discord.js");
const paginationEmbed = require("../../utils/paginationEmbed");
const guildDataModel = require("../../database/models/guildData");
const GameDocument = require("../../database/documents/game");

module.exports = {

    name: "leaderboard",
    aliases: ["lb", "top"],
    description: "Shows points leaderboard and win games of server." +
        "`gg lb` : It'll show all time leaderboard.\n" +
        "`gg lb daily` : It'll show today's leaderboard, if there are any winners.\n" +
        "`bb lb 2022-02-14` : Shows given date's leaderboard, if there were any winners." +
        "`gg lb 2022-02-14 2022-02-20` : Shows leaderboard b/w given dates, if there were any winners.\n",
    usage: "[daily | all | start-date] [end-date]",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

        let winners = [], title = "";
        if (args[0] === "daily") {
            winners = await getDailyLeaderboard(message.guild.id);
            title = "Today's Leaderboard";
        } else if (args[0] && !isNaN(Date.parse(args[0]))) {
            // calculate start and end date
            const startDate = new Date(args[0]);
            const nextDay = new Date(startDate.getTime() + 86400000);
            const endDate = new Date(args[1] || nextDay) || nextDay;
            // compare with min and max dates
            const minDate = new Date("2022-02-14");
            const maxDate = new Date(Date.now() + 86400000);
            if (startDate < minDate || startDate > maxDate || endDate < minDate || endDate > maxDate) {
                return message.channel.send({
                    content: `${client.myEmojis[1]} | **Invalid date**.\n` +
                        `Please use a date between <t:${minDate.getTime().toString().slice(0, -3)}> and <t:${maxDate.getTime().toString().slice(0, -3)}>`
                });
            }
            winners = await getSpecificDatesLeaderboard(message.guild.id, startDate, endDate);
            title = `Leaderboard of ${args[0]}`;
        } else {
            winners = await getAllTimeLeaderboard(message.guild.id);
            title = "All Time Leaderboard";
        }
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
                .setTitle(title)
                .setColor(client.colors[0])
                .setAuthor({
                    name: `${message.guild.name} has ${winners.length} Winnners`,
                    iconURL: client.user.displayAvatarURL({ format: "png", dynamic: true })
                })
                .setDescription(`\n${userDesc[i].substring(1)}` +
                    `[Please vote the bot! It helps a lot](https://top.gg/bot/${client.user.id}/vote)`)
        }
        return paginationEmbed(message, userEmbeds, emojiList, 30000);
    },
};

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

async function getDailyLeaderboard(guildID) {
    const games = await GameDocument.getDailyLeaderboard(guildID);
    if (!games || !games.length) {
        return false;
    }
    return getWinnersDataFromGames(games);
}

async function getSpecificDatesLeaderboard(guildID, startDate, endDate) {
    if (isNaN(Date.parse(startDate)) || isNaN(Date.parse(endDate))) {
        return false;
    }
    const games = await GameDocument.getGamesBetweenDates(guildID, startDate, endDate);
    if (!games || !games.length) {
        return false;
    }
    return getWinnersDataFromGames(games);
}

function getWinnersDataFromGames(games) {
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

    // format winners
    const winners = [];
    for (const userID in winnersData) {
        const data = winnersData[userID];
        winners.push(`<@${userID}> [**${data.wins}x**] with **${data.points}** Points\n`);
    }

    return winners;
}
