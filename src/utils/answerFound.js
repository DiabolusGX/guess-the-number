const guildConfigModel = require("../database/models/guildConfig");
const guildDataModel = require("../database/models/guildData");
const find = require("./find");

module.exports = async (client, message) => {
    const guildConfig = await guildConfigModel.findOne({ id: message.guild.id }).catch(console.error);

    if (guildConfig.reqRole && !message.member.roles.cache.has(guildConfig.reqRole)) return;

    const guildData = await guildDataModel.findOne({ id: message.guild.id }).catch(console.error);

    const guesses = client.games.get(message.channel.id).guesses;
    client.games.delete(message.channel.id);
    const games = guildData.runningGames;
    const { answer, points } = games.get(message.channel.id);
    games.delete(message.channel.id);

    /*let replyMsg = guildConfig.msg;

    const varMap = new Map();
    varMap.set("{user}", message.author);
    varMap.set("{answer}", answer);
    varMap.set("{points}", points);
    varMap.forEach((value, key) => {
        while (replyMsg.includes(key)) {
            replyMsg = replyMsg.replace(key, value);
        }
    });*/

    const users = guildData.users;
    if (users.has(message.author.id)) {
        const { wins, points } = users.get(message.author.id);
        users.set(message.author.id, { wins: wins + 1, points: points + 1 });
    }
    else users.set(message.author.id, { wins: 1, points: points });
    await guildDataModel.updateOne({ id: message.guild.id }, { $set: { runningGames: games, users: users } }, (err) => console.error);

    let reward = "";
    if (guildConfig.winRole) {
        const awardRole = await find.getRole(message, guildConfig.winRole);
        reward = `and **${awardRole.name}** role.`;
        message.member.roles.add(guildConfig.winRole, "Auto Role number guess").catch(console.error);
    }

    if (guildConfig.dm) {
        message.author.send(`${client.myEmojis[0]} | **Congratulations** 🎉\nYou guessed the correct number ||${answer}|| in ${message.channel} and won **${points} Points!**`);
    }
    message.channel.send(`${client.myEmojis[0]} | **Congratulations** ${message.author} 🎉\nYou guessed the correct number ||${answer}|| after total of **${guesses}** guesses!\n`+
        `Also, you received **${points} Points!** ${reward}`).then(msg => {
            if (msg.pinnable) msg.pin({ reason: "win message" }).catch(console.error);
        }).catch(console.error);

    if (guildConfig.lockRole) {
        const role = await find.getRole(message, guildConfig.lockRole);
        message.channel.updateOverwrite(role, { SEND_MESSAGES: false })
            .then(channel => {
                if(!channel.name.endsWith("🔒")) channel.edit({ name: channel.name + " 🔒" });
            })
            .catch(console.error);
    }

    return message.channel.messages.fetchPinned()
        .then(messages => {
            messages.forEach(m => {
                if(m.author.id === "818420448131285012" && !m.content.startsWith(`${client.myEmojis[0]} | **Congratulations**`))
                m.unpin({ reason: "game ended" }).catch(console.error);
            })
        })
        .catch(console.error);

}