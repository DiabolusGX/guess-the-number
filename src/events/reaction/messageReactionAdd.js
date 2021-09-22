
module.exports = async (_client, reaction, user) => {

    if (user.bot) return;
    if (reaction.message.partial) await reaction.message.fetch();
    if (reaction.partial) await reaction.fetch();

    if (!reaction.message.guild) return;

    if (reaction.message.channel.id === "889508466065043456" && reaction.message.id === "890276208447741992") {
        if (reaction.emoji.name === "🎉") {
            await reaction.message.guild.members.cache.get(user.id).roles.add("886974051946467328");
        }
    }

};