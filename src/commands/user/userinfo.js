const { MessageEmbed } = require("discord.js");
const guildDataModel = require("../../database/models/guildData");

module.exports = {

    name: "userinfo",
    aliases: ["info", "ui"],
    description: "Shows user's win games, points and other basic info.",

    async run(client, message, args) {

        let target;
        if (!args[0]) target = message.member;
        else {
            target = message.mentions.members.first() || message.guild.members.resolve(args[0]);
            if (!target) target = await message.guild.members.fetch(args[0]).catch(err => {
                return message.channel.send({
                    content: client.myEmojis[1] + " | **No user found!**"
                });
            });
        }
        if (!target) return;

        const dataDoc = await guildDataModel.findOne({ id: message.guild.id }).catch(console.error);
        const users = dataDoc.users;
        const userData = users.get(target.user.id);
        if (!userData) return message.channel.send({
            content: client.myEmojis[1] + " | **User has not won any game yet!**"
        });

        const userInfoEmbed = new MessageEmbed()
            .setColor(client.colors[0])
            .setAuthor({
                name: message.member.nickname || message.author.username,
                iconURL: message.author.displayAvatarURL({ format: "png", dynamic: true })
            })
            .setDescription(`${target} **${target.user.tag}** (\`${target.user.id}\`)\n\n` +
                `Total Wins : **${userData.wins}**⁣ 🎉\n` +
                `Total Points : **${userData.points}**⁣ ⚖️\n\n` +
                `[Please vote the bot! It helps a lot](https://top.gg/bot/${client.user.id}/vote)`)
            .setThumbnail(target.user.displayAvatarURL({ format: "png", dynamic: true, size: 1024 }))

        return message.channel.send({ embeds: [userInfoEmbed] });
    },
};