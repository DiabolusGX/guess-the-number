const { MessageEmbed } = require("discord.js");
const guildConfigModel = require("../../database/models/guildConfig");

module.exports = async (client, guild) => {

    if (!guild.available) return;

    guildConfigModel.deleteOne({ id: guild.id }, (err) => console.error).then(() => {
        console.log(`Deleted GUILD CONFIG for : "${guild.name}" owned by : "${ownerTag}"`);
    }).catch(console.error);

    const guildDeleteEmbed = new MessageEmbed()
        .setColor("#ff0000")
        .setAuthor({ name: ownerTag, iconURL: ownerAv })
        .setDescription(`**${guild.name}** (${guild.id}) by : ${guild.ownerId ? "<@" + guild.ownerId + ">" : null}, (\`${guild.ownerId}\`)`)
        .setThumbnail(guild.iconURL({ format: "png", dynamic: true, size: 1024 }) || ownerAv)
        .addFields(
            {
                name: `**Guild details**`, value: `**Guild memebrs** : ${guild.memberCount}` +
                    `**\nCreated at** : ${guild.createdAt}`, inline: true
            },
            {
                name: `**Other info**`, value: `**Total Guilds** :  ${client.guilds.cache.size}` +
                    `\n**Channels** : ${client.channels.cache.size} \n**Users** : ${client.users.cache.size}`, inline: true
            },
        )

    const joinLeaveChannel = await client.channels.cache.get("818439901044801567");
    return joinLeaveChannel.send(guildDeleteEmbed).catch(err => { if (err) console.log(err) });
}