const database = require("../../database/database");
const configDoc = require("../../utils/configDoc");
const dataDoc = require("../../utils/dataDoc");

const activities_list = [
    "your guesses || gg help",
    "birthday wishes for ${dev} || Join server to wish || gg invite for link",
    "join DEV Studios & wish ${dev} happy birthday || gg invite for link",
    "HAPPY BIRTHDAY ${dev}",
    "wishes for ${dev} on his birthday || gg invite for server link"
];

module.exports = async (client) => {
    await database();
    client.guilds.cache.forEach(async guild => {
        await configDoc(client, guild);
        const data = await dataDoc(client, guild);
        if (data.runningGames) data.runningGames.forEach((value, key) => client.games.set(key, { answer: value.answer, guesses: 0 }));
    });

    setInterval(() => {
        const index = Math.floor(Math.random() * (activities_list.length - 1) + 1);
        client.user.setActivity(activities_list[index], { type: "LISTENING" });
    }, 30000);
    //client.user.setActivity("your guesses || gg help", { type: "LISTENING" });
    console.log(client.user.tag + " is up and running in " + client.guilds.cache.size + " servers!");
}