const find = require("../../utils/find");

module.exports = {

    name: "hint",
    aliases: [],
    description: `To show hint on running games in any channel.\n` +
        `\`first\` - shows 1st digit of answer.\n` +
        `\`last\` - shows last digit of answer.\n` +
        `\`number\` - Tells if answer is smaller or greater then given *number*.`,
    category: "mod",
    guildOnly: true,
    usage: "<first | last | number> [channel]",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

        if (!args[0]) return message.channel.send(`${client.myEmojis[1]} | **No option provided!**\n` +
            `Please select valid option from - \`first\`, \`last\` or give a \`number\` to compare to.\n` +
            `Check \`${dbPrefix} help hint\` for more info.`);

        let targetChannel;
        if (args[1] || message.mentions.channels.first()) {
            targetChannel = await find.getChannel(message, args[2]);
        }
        else targetChannel = message.channel;
        if(!targetChannel) return message.reply(`${client.myEmojis[1]} | **Please mention valid channel**`);

        if (!client.games.has(targetChannel.id)) return message.channel.send(`${client.myEmojis[1]} | **No game going on in** ${targetChannel}`);
        const answer = client.games.get(targetChannel.id).answer;

        let hint;
        if (args[0] === "first") {
            const first = answer.toString()[0];
            hint = `First digit of answer is ||${first}||`;
        }
        else if (args[0] === "last") {
            const last = answer.toString().split("").pop();
            hint = `Last digit of answer is ||${last}||`;
        }
        else if (!isNaN(args[0])) {
            hint = answer > parseInt(args[0]) ? `${args[0]} is smaller then the actual answer`
                : `${args[0]} is greater then the actual answer`;
        }
        else return message.channel.send(`${client.myEmojis[1]} | **No option provided!**\n` +
            `Please select valid option from - \`first\`, \`last\` or give a \`number\` to compare to.\n` +
            `Check \`${dbPrefix} help hint\` for more info.`);

        return targetChannel.send(`${client.myEmojis[0]} | ${hint}. GL HF!!!`).then(msg => {
            if (msg.pinnable) msg.pin({ reason: "Game Hint" }).catch(console.error);
            return message.reply(`${client.myEmojis[0]} | Sent Hint in ${targetChannel} and pinned it!`).then(m => {
                setTimeout(() => { m.delete(); }, 5000);
            });
        });
    }
}