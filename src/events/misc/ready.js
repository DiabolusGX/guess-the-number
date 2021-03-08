const database = require("../../database/database");
const configDoc = require("../../utils/configDoc");
const dataDoc = require("../../utils/dataDoc");

module.exports = async (client) => {
    await database();
    client.guilds.cache.forEach(async guild => {
        await configDoc(client, guild);
        const data = await dataDoc(client, guild);
        if (data.runningGames) data.runningGames.forEach((value, key) => client.games.set(key, value.answer));
    });

    client.user.setActivity("Your Guesses || gg help", { type: "LISTENING" });
    console.log(client.user.tag + " is up and running in " + client.guilds.cache.size + " servers!");
}