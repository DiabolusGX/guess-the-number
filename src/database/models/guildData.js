const mongoose = require("mongoose");

const guildDataSchema = new mongoose.Schema({
    id: { type: String, unique: true },
    totalGames: { type: Number },
    totalWinners: { type: Number },
    runningGames: { type: Map },
    users: { type: Map },
    /* key = userId, value = {  }*/
});

const guildDataModel = module.exports = mongoose.model("guildData", guildDataSchema);