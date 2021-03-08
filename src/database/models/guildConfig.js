const mongoose = require("mongoose");

const guildConfigSchema = new mongoose.Schema({
    id: { type: String, unique: true },
    prefix: { type: String },
    premium: { type: Boolean },
    botManager: { type: String },
    dm: { type: Boolean },
    msg: { type: String },
    winRole: { type: String },
    reqRole: { type: String },
    lockRole: { type: String }
});

const guildConfigModel = module.exports = mongoose.model("guildConfig", guildConfigSchema);