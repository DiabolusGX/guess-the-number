const guildDataModel = require("../database/models/guildData");

module.exports = async (client, guild) => {
    const guildDataDoc = await guildDataModel.findOne({ id: guild.id }).catch(console.error);

    if (!guildDataDoc) {
        const dbGuildData = new guildDataModel({
            id: guild.id,
            totalGames: 0,
            runningGames: new Map(),
            users: new Map()
        });
        await dbGuildData.save().catch(console.error);
        return dbGuildData;
    }
    else return guildDataDoc;
}