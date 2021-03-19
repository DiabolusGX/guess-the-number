const { MessageEmbed } = require("discord.js");
const guildDataModel = require("../../database/models/guildData");

module.exports = {

    name: "userinfo",
    aliases: ["info"],
    description: "Shows user's win games, points and other basic info.",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

        let target;
        if(!args[0]) target = message.member;
        else{
            target = message.mentions.members.first() || message.guild.members.resolve(args[0]);
            if(!target) target = await message.guild.members.fetch(args[0]).catch(err => {
                return message.channel.send(client.myEmojis[1] + " | **No user found!**");
            });
        }
        if(!target) return;

        const dataDoc = await guildDataModel.findOne({ id: message.guild.id }).catch(console.error);
        const users = dataDoc.users;
        const userData = users.get(target.user.id);
        if(!userData) return message.channel.send(client.myEmojis[1] + " | **User has not won any game yet!**");

        const userInfoEmbed = new MessageEmbed()
            .setColor(client.colors[0])
            .setAuthor(message.member.nickname||message.author.username, message.author.displayAvatarURL({ format: "png", dynamic: true }))
            .setDescription(`${target} **${target.user.tag}** (\`${target.user.id}\`)\n\n`+
                `Total Wins : **${userData.wins}** 🎉\n`+
                `Total Points : **${userData.points}** ⚖️`)
            .setThumbnail( target.user.displayAvatarURL({ format: "png", dynamic: true, size: 1024 }) )
            .setFooter("Stats are maintained after user wins any game!")
        
        return message.channel.send(userInfoEmbed);
    },
};