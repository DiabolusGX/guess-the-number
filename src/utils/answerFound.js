const guildConfigModel = require("../database/models/guildConfig");
const guildDataModel = require("../database/models/guildData");
const GameDocument = require("../database/documents/game");
const find = require("./find");

module.exports = async (client, message) => {
    const guildConfig = await guildConfigModel.findOne({ id: message.guild.id }).catch(console.error);

    if (guildConfig.reqRole && !message.member.roles.cache.has(guildConfig.reqRole)) return;

    const guildData = await guildDataModel.findOne({ id: message.guild.id }).catch(console.error);

    // cache delete after fetching required info
    const gameData = client.games.get(message.channel.id);
    const { gameID, guesses } = gameData;
    client.games.delete(message.channel.id);

    // update games data to put in DB
    const games = guildData.runningGames;
    const { answer, points } = games.get(message.channel.id);
    games.delete(message.channel.id);

    // update users data to put in DB
    const users = guildData.users;
    if (users.has(message.author.id)) {
        const userData = users.get(message.author.id);
        const wins = userData.wins;
        const userPoints = userData.points;
        users.set(message.author.id, { wins: wins + 1, points: userPoints + 1 });
    }
    else users.set(message.author.id, { wins: 1, points: points });

    // update DB
    await guildDataModel.updateOne({ id: message.guild.id }, { $set: { runningGames: games, users: users } }, (err) => console.error);
    await GameDocument.finishGame(gameID, message.author.id, points, guesses);

    // post result
    let reward = "";
    if (guildConfig.winRole) {
        const awardRole = await find.getRole(message, guildConfig.winRole);
        reward = `and **${awardRole}** role.`;
        message.member.roles.add(guildConfig.winRole, "Auto Role number guess").catch(console.error);
    }

    if (guildConfig.dm) {
        message.author.send({
            content: `${client.myEmojis[0]} | **Congratulations** 🎉\nYou guessed the correct number ||${answer}|| in ${message.channel} and won **${points} Points!**\n` +
                `Please vote the bot! It helps a lot <https://top.gg/bot/${client.user.id}/vote>`
        });
    }
    message.channel.send({
        content: `${client.myEmojis[0]} | **Congratulations** ${message.author} 🎉\nYou guessed the correct number ||${answer}|| after total of **${guesses}** guesses!\n` +
            `Also, you received **${points} Points!** ${reward}\n` +
            `Please vote the bot! It helps a lot <https://top.gg/bot/${client.user.id}/vote>`
    }).then(msg => {
        if (msg.pinnable) msg.pin({ reason: "win message" }).catch(console.error);
    }).catch(console.error);

    if (guildConfig.lockRole) {
        const role = await find.getRole(message, guildConfig.lockRole);
        message.channel.permissionOverwrites.edit(
            role,
            { SEND_MESSAGES: true },
            { reason: "Game ended, locking channel for lock role", type: 0 }
        )
            .then(channel => {
                if (!channel.name.endsWith("🔒")) channel.edit({ name: channel.name + " 🔒" });
            })
            .catch(console.error);
    }

    return message.channel.messages.fetchPinned()
        .then(messages => {
            messages.forEach(m => {
                if (m.author.id === client.user.id && !m.content.startsWith(`${client.myEmojis[0]} | **Congratulations**`)) {
                    m.unpin({ reason: "game ended" }).catch(console.error);
                }
            })
        })
        .catch(console.error);

}