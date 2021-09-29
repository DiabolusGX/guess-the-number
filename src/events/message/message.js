const { MessageEmbed } = require("discord.js");
const guildConfigModel = require("../../database/models/guildConfig");
const answerFound = require("../../utils/answerFound");

const fs = require("fs");
const scoresData = require("../../../scores.json");

module.exports = async (client, message) => {

    if (message.author.bot || message.channel.type !== "text") return;

    if (!isNaN(message.content) && client.games.has(message.channel.id)) {
        if (parseInt(message.content) === client.games.get(message.channel.id).answer) return answerFound(client, message);
        else {
            const { answer, guesses } = client.games.get(message.channel.id);
            client.games.set(message.channel.id, { answer: answer, guesses: guesses + 1 });
        }
    }

    const dbPrefix = await client.guildConfigPrefix.get(message.guild.id) || process.env.PREFIX;
    const botMentioned = message.content.search(/<@!818420448131285012>/i);
    if (botMentioned >= 0) {
        const prefixEmbed = new MessageEmbed()
            .setColor(client.colors[0])
            .setDescription(`🥳 My prefix for **${message.guild.name}** is : **\`${dbPrefix}\`**`);
        return message.channel.send(prefixEmbed).then(msg => setTimeout(() => msg.delete(), 5000));
    }

    if (!message.content.startsWith(dbPrefix)) return;
    const args = message.content.slice(dbPrefix.length).split(/ +/);
    let cmdName = args.shift()?.toLowerCase();
    if (!cmdName) cmdName = args.shift()?.toLowerCase();
    if (!cmdName) return;

    const command = client.commands.get(cmdName);
    if (!command) return;

    if (message.author.id !== "454611998051794954") {
        if (command.category === "owner") return;
        if (command.category === "mod" && !message.member.permissions.has("ADMINISTRATOR")) {
            const guildConfigDoc = await guildConfigModel.findOne({ id: message.guild.id }).catch(console.error);
            if (guildConfigDoc.botManager) {
                if (!message.member.roles.cache.has(guildConfigDoc.botManager)) return message.reply("Only Admins or Bot Manager can use `" + command.name + "`.").then(msg => setTimeout(() => msg.delete(), 5000));
            }
            else return message.reply("Only Admins or Bot Manager can use `" + command.name + "`.").then(msg => setTimeout(() => msg.delete(), 5000));
        }
    }

    if (command.guildOnly && message.channel.type !== "text") return;
    if (command.args && !args.length) {
        let reply = `You didn't provide any arguments, ${message.author}!`;
        if (command.usage) {
            reply += `\nThe proper usage would be: \`${dbPrefix}${command.name} ${command.usage}\``;
        }
        return message.channel.send(reply).then(msg => setTimeout(() => msg.delete(), 5000));
    }
    const cooldowns = client.cooldowns;
    if (!cooldowns.has(command.name)) cooldowns.set(command.name, new Map());

    const now = Date.now();
    const timestamps = cooldowns.get(command.name);
    const cooldownAmount = (command.cooldown || 3) * 1000;

    if (timestamps.has(message.author.id)) {
        const expirationTime = timestamps.get(message.author.id) + cooldownAmount;
        if (now < expirationTime) {
            const timeLeft = (expirationTime - now) / 1000;
            let d = Number(timeLeft);
            let h = Math.floor(d / 3600);
            let m = Math.floor(d % 3600 / 60);
            let s = Math.floor(d % 3600 % 60);

            let hDisplay, mDisplay, sDisplay, time;
            if (h > 0) { hDisplay = h + " Hour(s), "; } else hDisplay = "";
            if (m > 0) { mDisplay = m + " Minute(s), "; } else mDisplay = "";
            if (s >= 0) { sDisplay = s + " Second(s)"; } else sDisplay = "";
            time = hDisplay + mDisplay + sDisplay;
            return message.reply(`Please wait \`${time}\` before reusing the \`${command.name}\` command again.`).then(msg => setTimeout(() => msg.delete(), 5000));
        }
    }
    timestamps.set(message.author.id, now);
    setTimeout(() => timestamps.delete(message.author.id), cooldownAmount);

    try {
        await command.run(client, message, args);
    } catch (error) {
        return console.error(message.content, message.guild.id, message.guild.name, error);
    }
}