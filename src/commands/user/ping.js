const { MessageEmbed } = require("discord.js");

module.exports = {

    name : "ping",
    aliases: [],
    description : "Shows bot ping and response time!",
    category: "user",
    guildOnly: true,
    cooldown: 10,

    async run(client, message, args){

        let firstMsg, responseTime;
        await message.channel.send('⚙️ | Calculating...').then( msg =>{
            if(msg) firstMsg = msg;
            responseTime = msg.createdTimestamp;
        })
        responseTime = responseTime - message.createdTimestamp;

        const pingEmbed = new MessageEmbed()
            .setColor(client.colors[0])
            .setTitle(`🏓 Pong!`)
            .setDescription(`Ping is : ${Math.round(message.client.ws.ping)} ms.\nResponse time : ${responseTime} ms!`)
            .setFooter(`Made with ❤️ by DiabolusGX`, `${client.user.displayAvatarURL({ format: "png", dynamic: true })}`);
        
        return firstMsg.edit("", pingEmbed);
    },
};