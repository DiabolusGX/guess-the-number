const mongoose = require("mongoose");

const Game = new mongoose.Schema({
    guildID: { type: String, index: true },
    channelID: { type: String, index: true },
    createdBy: { type: String },
    wonBy: { type: String },
    answer: { type: Number },
    points: { type: Number, default: 0 },
    guesses: { type: Number, default: 0 },
    finishedAt: { type: Date, index: true },
    finished: { type: Boolean, default: false },
    disabled: { type: Boolean, default: false },
    createdAt: { type: Date, default: Date.now },
    updatedAt: { type: Date, default: Date.now },
});

module.exports = mongoose.model("game", Game);
