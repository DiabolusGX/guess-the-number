const { Guild } = require("discord.js");

module.exports = {

    /**
     * @param  {Message} message message reference
     * @param  {String} name Role Name OR id
     */
    getRole: (message, name) => {
        if (!name) return null;
        let role;
        if (message.mentions.roles.first()) {
            role = message.mentions.roles.first();
        }
        else role = message.guild.roles.cache.find(r => r.name.toLowerCase() === name.toLowerCase()) || message.guild.roles.cache.get(name);
        return role;
    },

    /**
     * @param {Message} message message reference
     * @param {String} name TEXT Channel Name OR id
     */
    getChannel: (message, name) => {
        let channel;
        if (message.mentions.channels.first()) {
            channel = message.mentions.channels.first();
        }
        if (!channel) channel = message.guild.channels.cache.find(c => c.id === name) ||
            message.guild.channels.cache.get(name);
        if (!channel) channel = message.guild.channels.cache.find(c => c.name.toLowerCase() === name.toLowerCase()) ||
            message.guild.channels.cache.get(name);

        if (channel?.type !== "GUILD_TEXT" && channel?.type !== "GUILD_NEWS") return null;
        else return channel;
    },

    /**
     * @param {Guild} guild Target Guild reference
     * @param {String} name VOICE Channel Name OR id
     */
    fetchChannel: (guild, name) => {
        return guild.channels.cache.find(c => c.name === name) ||
            guild.channels.cache.get(name);
    }

}