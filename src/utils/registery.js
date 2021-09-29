const fs = require("fs").promises;
const path = require("path");

async function registerCommands(client, dir) {
    let files = await fs.readdir(path.join(__dirname, dir));
    // Loop through each file.
    files.forEach(async file => {
        let stat = await fs.lstat(path.join(__dirname, dir, file));
        if (stat.isDirectory()) // If file is a directory, recursive call recurDir
            registerCommands(client, path.join(dir, file));
        else {
            // Check if file is a .js file.
            if (file.endsWith(".js")) {
                let cmdName = file.substring(0, file.indexOf(".js"));
                let cmdModule = require(path.join(__dirname, dir, file));
                client.commands.set(cmdName, cmdModule);
                if (cmdModule.aliases && cmdModule.aliases.length !== 0) cmdModule.aliases.forEach(alias => client.commands.set(alias, cmdModule));
            }
        }
    });
}

async function registerEvents(client, dir) {
    let files = await fs.readdir(path.join(__dirname, dir));
    // Loop through each file.
    files.forEach(async file => {
        let stat = await fs.lstat(path.join(__dirname, dir, file));
        if (stat.isDirectory()) // If file is a directory, recursive call recurDir
            registerEvents(client, path.join(dir, file));
        else {
            // Check if file is a .js file.
            if (file.endsWith(".js")) {
                let eventName = file.substring(0, file.indexOf(".js"));
                let eventModule = require(path.join(__dirname, dir, file));
                client.on(eventName, eventModule.bind(null, client));
            }
        }
    });
}

module.exports = {
    registerCommands,
    registerEvents
}
