const guildConfigModel = require("../database/models/guildConfig");
const guildDataModel = require("../database/models/guildData");
const find = require("./find");

module.exports = async (client, message) => {
    const guildConfig = await guildConfigModel.findOne({ id: message.guild.id }).catch(console.error);

    if (guildConfig.reqRole && !message.member.roles.cache.has(guildConfig.reqRole)) return;

    const guildData = await guildDataModel.findOne({ id: message.guild.id }).catch(console.error);

    client.games.delete(message.channel.id);
    const games = guildData.runningGames;
    const { answer, points } = games.get(message.channel.id);
    games.delete(message.channel.id);

    let replyMsg = guildConfig.msg;

    const varMap = new Map();
    varMap.set("{user}", message.author);
    varMap.set("{answer}", answer);
    varMap.set("{points}", points);
    varMap.forEach((value, key) => {
        while (replyMsg.includes(key)) {
            replyMsg = replyMsg.replace(key, value);
        }
    });

    const users = guildData.users;
    if(users.has(message.author.id)) {
        const { wins, points } = users.get(message.author.id);
        users.set(message.author.id, { wins:  wins+1, points: points+1 });
    }
    else users.set(message.author.id, { wins: 1, points: points });
    await guildDataModel.updateOne({ id: message.guild.id }, { $set: { runningGames: games, users: users } }, (err) => console.error);

    if (guildConfig.winRole) message.member.roles.add(guildConfig.winRole, "Auto Role number guess").catch(console.error);

    if (guildConfig.dm) {
        message.author.send(`${client.myEmojis[0]} | **Congratulations** 🎉\nYou guessed the correct number ||${answer}|| in ${message.channel} and won **${points} Points!**`);
    }
    message.channel.send(`${client.myEmojis[0]} | **Congratulations** ${message.author} 🎉\nYou guessed the correct number ||${answer}|| and won **${points} Points!**`)
        .then(msg => msg.pin({ reason: "win message" }).catch(console.error))
        .catch(console.error)

    const role = await find.getRole(message, guildConfig.lockRole);
    return message.channel.updateOverwrite(role, { SEND_MESSAGES: false })
        .then(channel => channel.edit({ name: channel.name + " 🔒" }))
        .catch(console.error);
}