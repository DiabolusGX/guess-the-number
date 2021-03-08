require("dotenv").config();
const { Client } = require("discord.js");
// const DBL = require("dblapi.js");
// const dbl = new DBL(process.env.DBL_TOKEN, { webhookPort: 8081, webhookAuth: "king.Dbl@07" });
const client = new Client();
const registery = require("./utils/registery");

(async () => {
    client.login(process.env.BOT_TOKEN);
    client.commands = new Map();
    client.cooldowns = new Map();
    client.guildConfigPrefix = new Map();
    client.colors = ["#2f3136"]
    await registery.registerCommands(client, "../commands");
    await registery.registerEvents(client, "../events");
})();

// dbl.webhook.on("ready", webhook => {
//     const webhookReadyEvent = require("./webhook/ready");
//     return webhookReadyEvent(webhook, client, dbl);
// });
// dbl.webhook.on("vote", vote => {
//     const voteEvent = require("./webhook/vote");
//     return voteEvent(vote, client);
// });