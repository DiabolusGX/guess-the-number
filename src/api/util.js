const path = require("path");
require("dotenv").config({ path: path.join(__dirname, "/../../.env") });
const axios = require("axios");
const baseURL = process.env.API_URL;
const token = process.env.BB_TOKEN;

const getGuildPremiumInfo = async (guildID) => {
    try {
        const result = await axios({
            method: "get",
            url: `${baseURL}/premium/guild-info/${guildID}`,
            headers: {
                "Authorization": `Bot ${token}`
            }
        });
        const data = result.data;
        const isPremium = data.premium ?? false;
        return { isPremium, bbAdded: true };
    }
    catch (err) {
        if (err.response.status === 404) {
            return { isPremium: false, bbAdded: false };
        }
        return { isPremium: false, bbAdded: true };
    };
}

const getUserPremiumInfo = async (userID) => {
    try {
        const result = await axios({
            method: "get",
            url: `${baseURL}/premium/user-info/${userID}`,
            headers: {
                "Authorization": `Bot ${token}`
            }
        })
        const data = result.data;
        const { premiumGuilds, premiumRemaining } = data;
        return {
            premiumGuilds,
            premiumRemaining,
            isUserAvailable: true
        };
    }
    catch (err) {
        if (err.response.status === 404) {
            return {
                premiumGuilds: [],
                premiumRemaining: 0,
                isUserAvailable: false,
            };
        }
        return {
            premiumGuilds: [],
            premiumRemaining: 0,
            isUserAvailable: true,
        };
    }
}

module.exports = {
    getGuildPremiumInfo,
    getUserPremiumInfo
}
