// MongoDB initialization script for Guess The Number bot

// Switch to the bot database
db = db.getSiblingDB(process.env.MONGO_INITDB_DATABASE || "guess_the_number");

// Create collections with initial indexes
print("Initializing Guess The Number database...");

// Games collection
db.createCollection("games");
db.games.createIndex({ guild_id: 1, channel_id: 1 });
db.games.createIndex({ created_at: 1 }, { expireAfterSeconds: 86400 }); // TTL index for cleanup

// Game stats collection
db.createCollection("game_stats");
db.game_stats.createIndex({ user_id: 1, guild_id: 1 });
db.game_stats.createIndex({ guild_id: 1 });

// Guild configurations collection
db.createCollection("guild_configs");
db.guild_configs.createIndex({ guild_id: 1 }, { unique: true });

// Guild data collection
db.createCollection("guild_data");
db.guild_data.createIndex({ guild_id: 1 }, { unique: true });

print("Database initialization completed successfully!");
print("Collections created: games, game_stats, guild_configs, guild_data");

// Create a test user for validation (optional)
if (process.env.NODE_ENV === "development") {
    print("Development mode: Creating sample data...");

    // Insert sample guild config
    db.guild_configs.insertOne({
        id: "709339522680356914",
        prefix: "gg",
        dm: true,
        winRole: "718919390359322644",
        lockRole: "748504332785156126",
        logChannel: "744633854253334548",
        autoReactionHints: true,
    });

    // add more sample data here...

    print("Sample data created for development environment");
}
