const mongoose = require("mongoose");

const guildDataSchema = new mongoose.Schema({
    id: { type: String, unique: true },
    totalGames: { type: Number },
    runningGames: { type: Map },
    /* key = channelId, value = { answer: Number, points: Number } */
    users: { type: Map },
    /* key = userId, value = { wins: Number, points: Number }*/
});

const guildDataModel = module.exports = mongoose.model("guildData", guildDataSchema);