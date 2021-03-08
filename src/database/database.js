require("dotenv").config();
const mongoose = require("mongoose");

module.exports = async () => {
    mongoose.set("useFindAndModify", false);
    mongoose.set("useCreateIndex", true);
    mongoose.connect(process.env.MONGO_URL, { dbName: "guessTheNumber", useNewUrlParser: true, useUnifiedTopology: true })
        .then(() => console.log("connected to mongodb."));
}