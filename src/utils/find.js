module.exports = {

    async getRole(message, args) {
        let role;
        if (message.mentions.roles.first()) {
            role = message.mentions.roles.first();
        }
        else role = message.guild.roles.cache.find(role => role.name.toLowerCase() === args.toLowerCase()) || message.guild.roles.cache.get(args);
        return role;
    },

    async getChannel(message, args) {
        let channel;
        if (message.mentions.channels.first()) {
            channel = message.mentions.channels.first();
        }
        else channel = message.guild.channels.cache.find(channel => channel.name === args) ||
            message.guild.channels.cache.get(args);
        if (!channel || channel.type !== "text") return null;
        else return channel;
    },

    async fetchChannel(guild, args) {
        const channel = guild.channels.cache.find(channel => channel.name === args) ||
            guild.channels.cache.get(args);
        return channel;
    }

}