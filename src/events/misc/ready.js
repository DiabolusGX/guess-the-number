const database = require("../../database/database");
const configDoc = require("../../utils/configDoc");

module.exports = async (client) => {
    await database();
    client.guilds.cache.forEach(async guild => {
        await configDoc(client, guild);
    });

    client.user.setActivity("numbers", { type: "LISTENING" });
    console.log(client.user.tag + " is up and running in " + client.guilds.cache.size + " servers!");
}