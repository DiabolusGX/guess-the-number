const path = require("path");
require("dotenv").config({ path: path.join(__dirname, "/../.env") });
const { Client, Intents, Options } = require("discord.js");
const client = new Client({
    shards: "auto",
    partials: ["MESSAGE", "REACTION"],
    makeCache: Options.cacheWithLimits({
        ThreadManager: 0,
        MessageManager: 0,
        PresenceManager: 0,
        VoiceStateManager: 0,
        GuildEmojiManager: 0,
        GuildInviteManager: 0,
    }),
    intents: [
        Intents.FLAGS.GUILDS,
        Intents.FLAGS.GUILD_MEMBERS,
        Intents.FLAGS.GUILD_MESSAGES,
        Intents.FLAGS.GUILD_MESSAGE_REACTIONS,
    ],
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
