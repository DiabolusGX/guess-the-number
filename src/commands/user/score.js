const { MessageEmbed } = require("discord.js");
const scoresData = require("../../../scores.json");

module.exports = {

    name: "score",
    description: "Shows your score and leaderboard on basis of how many times you ping **Joey**.",

    async run(client, message, args) {

        if (message.author.id !== "454611998051794954") return;
        const authorId = message.author.id;
        const totalSpammers = Object.keys(scoresData).length;
        let desc = "", rank = "", first = "", second = "", third = "", totalPings = 0;

        Object.keys(scoresData).sort((a, b) => scoresData[b] - scoresData[a]).forEach((key, index) => {
            totalPings += scoresData[key];
            if (index === 0) first = `\`[1]\` | <@${key}> => ${scoresData[key]}`;
            if (index === 1) second = `\`[2]\` | <@${key}> => ${scoresData[key]}`;
            if (index === 2) third = `\`[3]\` | <@${key}> => ${scoresData[key]}`;
            if (key === authorId) rank = `Your rank is **${index + 1}** out of total **${totalSpammers}** spammers!`;
        });

        const count = scoresData.hasOwnProperty(`${authorId}`) ? scoresData[`${authorId}`] : 0;
        desc += `You've contributed **${count}** pings out of **${totalPings}** total pings!\n`;

        desc += `\nOur top 3 spammers are:\n`;
        desc += `${first}\n`;
        desc += `${second}\n`;
        desc += `${third}\n`;
        desc += `\n${rank}`;

        const scoreEmbed = new MessageEmbed()
            .setColor(client.colors[0])
            .setAuthor({
                name: `We have ${totalSpammers} Spammers`,
                iconURL: client.user.displayAvatarURL({ format: "png", dynamic: true })
            })
            .setDescription(desc)
            .setFooter({
                text: "Top spammer will get 200 points in Fun Week by Diabolus ❤️"
            })

        return message.channel.send({
            content: `Ping Scores 🔔`,
            embeds: [scoreEmbed]
        });
    },
};