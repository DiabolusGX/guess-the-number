const { MessageEmbed } = require("discord.js");
const configDoc = require("../../utils/configDoc");
const dataDoc = require("../../utils/dataDoc");

module.exports = async (client, guild) => {

    await configDoc(client, guild);
    await dataDoc(client, guild);

    let ownerId = "NA", ownerTag = "NA", owner = "NA", ownerAv = "https://i.imgur.com/X0CBc8U.gif";
    if (guild.ownerId) {
        ownerId = guild.ownerId;
        owner = await guild.members.fetch(ownerId).catch(() => console.error("err while fetching owner"));
        ownerTag = owner?.user?.tag;
        ownerAv = owner?.user.displayAvatarURL({ format: "png", dynamic: true, size: 1024 });
    }

    const guildCreateEmbed = new MessageEmbed()
        .setColor("#00ff00")
        .setAuthor({ name: ownerTag, iconURL: ownerAv })
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

    const joinLeaveChannel = await client.channels.cache.get("818439901044801567");
    return joinLeaveChannel.send({ embeds: [guildCreateEmbed] }).catch(console.error);
}