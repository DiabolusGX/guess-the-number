const guildConfigModel = require("../database/models/guildConfig");

module.exports = async (client, guild) => {
    const guildConfigDoc = await guildConfigModel.findOne({ id: guild.id }).catch(console.error);

    if (!guildConfigDoc) {
        const dbGuildConfig = new guildConfigModel({
            id: guild.id,
            prefix: "gg",
            dm: true,
            msg: "Thanks for participating everyone and Congratulations {user}, You've has guessed the currect number i.e {number}",
            lockRole: guild.roles.everyone.id,
        });
        await dbGuildConfig.save().catch(console.error);
        client.guildConfigPrefix.set(guild.id, "gg");
        return dbGuildConfig;
    }
    else {
        if (!client.guildConfigPrefix.has(guild.id)) client.guildConfigPrefix.set(guild.id, guildConfigDoc.prefix);
        return guildConfigDoc;
    }
}