require("dotenv").config();
const { Client } = require("discord.js");
const client = new Client({
    messageCacheMaxSize: 10,
    messageCacheLifetime: 10,
    messageSweepInterval: 60,
    partials: ["MESSAGE", "REACTION"],
    intents: ["GUILDS", "GUILD_MEMBERS", "GUILD_MESSAGES", "GUILD_MESSAGE_REACTIONS"],
    allowedMentions: { parse: ["users"], repliedUser: true }
});
const registery = require("./utils/registery");

(async () => {
    client.games = new Map();
    client.commands = new Map();
    client.cooldowns = new Map();
    client.guildConfigPrefix = new Map();
    client.colors = ["#2f3136"];
    client.myEmojis = ["<:greentick:768464483009691648>", "<:redtick:768464519638024233>"];
    await registery.registerCommands(client, "../commands");
    await registery.registerEvents(client, "../events");
    client.login(process.env.BOT_TOKEN);
})();

// Warning handling err catch
process.on("unhandledRejection", error => {
    console.error("Unhandled promise rejection:", error);
});