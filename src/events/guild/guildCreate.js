const { MessageEmbed } = require("discord.js");
const configDoc = require("../../utils/configDoc");

module.exports = async (client, guild) => {

    await configDoc(guild);

    let ownerId = "NA", ownerTag = "NA", owner = "NA", ownerAv = "https://i.imgur.com/X0CBc8U.gif";
    if (guild.owner) {
        owner = guild.owner.user; ownerId = guild.owner.id; ownerTag = guild.owner.user.tag; ownerAv = guild.owner.user.displayAvatarURL({ format: "png", dynamic: true, size: 1024 });
    }

    const guildCreateEmbed = new MessageEmbed()
        .setColor("#00ff00")
        .setAuthor(`${ownerTag}`, ownerAv)
        .setTitle(`Guild Created`)
        .setDescription(`**${guild.name}** (${guild.id}) by : ${owner}, ${ownerTag} (${ownerId})`)
        .setThumbnail(guild.iconURL({ format: "png", dynamic: true, size: 1024 }) || ownerAv)
        .addFields(
            {
                name: `**Guild details**`, value: `**Guild memebrs** : ${guild.memberCount}` +
                    `\n**Created at** : ${guild.createdAt}`, inline: true
            },
            {
                name: `**Other info**`, value: `**Total Guilds** :  ${client.guilds.cache.size}` +
                    `\n**Channels** : ${client.channels.cache.size} \n**Users :** ${client.users.cache.size}`, inline: true
            },
        )
        .setTimestamp()
        .setFooter(`Guild Created | ${client.user.username}`, `${client.user.displayAvatarURL({ format: "png", dynamic: true })}`);

    const joinLeaveChannel = await client.channels.cache.get("818439901044801567");
    return joinLeaveChannel.send(guildCreateEmbed).catch(console.error);
}