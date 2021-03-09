const guildDataModel = require("../../database/models/guildData");
const find = require("../../utils/find");

module.exports = {

    name: "start",
    aliases: ["go"],
    description: `Start the event with minimum and maximum number in a specific channel.`,
    category: "mod",
    guildOnly: true,
    usage: "<min> <max> [channel]",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);
        const dataDoc = await guildDataModel.findOne({ id: message.guild.id });

        if (!args[0] || !args[1] || isNaN(args[0]) || isNaN(args[1]) || parseInt(args[0]) > parseInt(args[1]))
            return message.channel.send(`${client.myEmojis[1]} | **Minimum and Maximum value missing!**\n` +
                `Min & Max value are important to start Guess The Number Game like : \n` +
                `- \`${dbPrefix} start 69 420\` will start game in current channel with random number b/w 69 and 420.\n` +
                `- \`${dbPrefix} start 100 1000 #channel\` will start game in mentioned channel with random number b/w 100 and 10000.\n`);

        const min = parseInt(args[0]);
        const max = parseInt(args[1]);
        const points = Math.floor( (max-min) / 10 );
        let targetChannel;
        if (args[2] || message.mentions.channels.first()) {
            targetChannel = await find.getChannel(message, args[2]);
        }
        else targetChannel = message.channel;
        if(client.games.has(targetChannel.id)) return message.channel.send(`${client.myEmojis[1]} | There is already a game going on in ${targetChannel}`);

        const randomAnswer = Math.floor(Math.random() * (max - min) + min);

        const totalGames = dataDoc.totalGames;
        const games = dataDoc.runningGames;
        games.set(targetChannel.id, {answer: randomAnswer, points: points });
        await guildDataModel.updateOne({ id: message.guild.id }, { $set: { runningGames: games, totalGames: totalGames+1 } }, (err) => console.error);
        client.games.set(targetChannel.id, randomAnswer);

        return message.channel.send(`${client.myEmojis[0]} | **Started Game** in ${targetChannel}`).then(msg => {
            targetChannel.send(`${client.myEmojis[0]} | **Started Game**\nMin number is \`${min}\` & Max number is \`${max}\``)
                .then(m => { if (m.pinnable) m.pin({ reason: "Game Start" }).catch(console.error); }).catch(console.error);
            return message.author.send(`${client.myEmojis[0]} | Started game with random answer ||${randomAnswer}|| in ${targetChannel}`)
                .catch(err => message.channel.send("Could not send game's answer to " + message.author));
        });
    }
}