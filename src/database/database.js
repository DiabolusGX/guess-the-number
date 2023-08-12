const path = require("path");
require("dotenv").config({ path: path.join(__dirname, "/../../.env") });
const mongoose = require("mongoose");

module.exports = async () => {
    mongoose.set("useFindAndModify", false);
    mongoose.set("useCreateIndex", true);
    return mongoose.connect(process.env.MONGO_URL, { dbName: "guessTheNumber", useNewUrlParser: true, useUnifiedTopology: true });
}