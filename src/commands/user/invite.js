const { MessageEmbed } = require("discord.js");

module.exports = {

    name: "invite",
    aliases: ["vote"],
    description: `It'll provide you useful links for the bot.`,

    async run(client, message, args) {

        const defaultInvite = `https://discord.com/api/oauth2/authorize?client_id=${client.user.id}&permissions=268790897&scope=bot`;
        const adminInvite = `https://discord.com/api/oauth2/authorize?client_id=${client.user.id}&permissions=8&scope=bot`;

        const inviteEmbed = new MessageEmbed()
            .setColor(client.colors[0])
            .setAuthor({
                name: "Bot Invite Links",
                iconURL: "https://cdn.discordapp.com/attachments/797404067432890378/797404132037361684/partying-face.png"
            })
            .setDescription(`\n**Bot invite** : \n[Default Invite](${defaultInvite})\n[Admin Invite](${adminInvite}) (recommended)` +
                `\n\n__Vote Link__ : \nhttps://top.gg/bot/${client.user.id}/vote\n` +
                `\n__Support Server__ : \nhttps://discord.gg/8kdx63YsDf\n`)
            .setFooter({
                text: "Checkout Booster Bot - Everything for server boosts",
                iconURL: "https://cdn.discordapp.com/attachments/699571659635556372/782257966933344256/622857534742593536.gif"
            })
        return message.channel.send({ embeds: [inviteEmbed] });
    },
};