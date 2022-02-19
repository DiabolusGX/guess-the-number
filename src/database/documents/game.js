const Game = require("../models/game");

module.exports = {
    /**
     * Create game event with all detailes while starting game.
     * @param {string} guildID server id
     * @param {string} channelID channel id
     * @param {string} createdBy user id
     * @param {number} points points
     * @returns Game
     */
    createGame: async (guildID, channelID, createdBy, points, answer) => {
        const game = new Game({
            guildID,
            channelID,
            createdBy,
            points,
            answer
        });
        await game.save();
        return game;
    },

    /**
     * Finish game event with all detailes while finishing game.
     * @param {string} gameID game event id
     * @param {string} wonBy user id
     * @param {number} points points
     */
    finishGame: async (gameID, wonBy, points, guesses) => {
        try {
            return Game.findOneAndUpdate({
                _id: gameID
            }, {
                wonBy,
                points,
                guesses,
                finished: true,
                finishedAt: new Date(),
                updatedAt: Date.now()
            }, {
                new: true
            });
        } catch (error) {
            console.error(error);
        }
    },

    /**
     * Get finished game events in last 24 hours.
     * @param {string} guildID server id
     * @returns {Promise<array>} array of games
     */
    getDailyLeaderboard: async (guildID) => {
        try {
            return Game.find({
                guildID,
                finished: true,
                finishedAt: {
                    $gte: new Date(Date.now() - 86400000)
                }
            }).sort({
                points: -1
            }).select({
                _id: 0,
                points: 1,
                wonBy: 1
            });
        } catch (error) {
            console.error(error);
        }
    },

    /**
     * Get all finished game events between 2 given dates.
     * @param {string} guildID server id
     * @param {Date} startDate start date
     * @param {Date} endDate end date
     * @returns {Promise<array>} array of games
     */
    getGamesBetweenDates: async (guildID, startDate, endDate) => {
        try {
            return Game.find({
                guildID,
                finished: true,
                finishedAt: {
                    $gte: new Date(startDate),
                    $lte: new Date(endDate)
                }
            }).sort({
                points: -1
            }).select({
                _id: 0,
                points: 1,
                wonBy: 1
            });
        } catch (error) {
            console.error(error);
        }
    },

    /**
     * Disable the game event.
     * @param {string} gameID game event id
     * @returns {Promise<boolean>} true if game event was disabled
     */
    disableGame: async (gameID, guesses) => {
        try {
            await Game.findOneAndUpdate({
                _id: gameID
            }, {
                guesses,
                disabled: true
            });
            return true;
        } catch (error) {
            console.error(error);
            return false;
        }
    },
};
