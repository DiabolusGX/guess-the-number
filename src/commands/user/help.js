const { MessageEmbed } = require("discord.js");

module.exports = {

    name : "help",
    aliases: ["commands"],
    description : "Shows all available commands.",
    category: "user",
    guildOnly: false,
    
    async run(client, message, args){

        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

        const discription = []; const moreInfo = []; const allCommandNames = [];
        const { commands } = message.client;

        if (!args.length) {
            for (let key of commands.keys()) allCommandNames.push(key);
            allCommandNames.join("` `");
            moreInfo.push(`\nYou can send \`${dbPrefix} help [command name]\` to get info on a specific command!`);
            discription.push(`\nTo get started with setting up Boost Tracking, Greeting and Logging - **__\`${dbPrefix} guide\`__**\n\n`);

            const generalHelpEmbed = new MessageEmbed()
                .setColor('#87339c')
                //.setAuthor(`${message.author.username}`, `${message.author.displayAvatarURL({ format: "png", dynamic: true })}`)
                .setDescription(`${discription}\n\n`)
                //.setThumbnail(message.guild.iconURL({ format: "png", dynamic: true }) || client.user.displayAvatarURL({ format: "png", dynamic: true }))
                .addFields(
                    { name: "\u200B", value: "_ _" },
                    { name: 'Setup Commands', value: `__For Admins and Bot Manager__\n\n`+
                        `\`setup\` : Setup Server settings.\n` +
                        `\`greet\` : Setup Greeting settings.\n` +
                        `\`rank\` : Add/Remove role to boosters\n` +
                        `\`emoji\` : Lock/unlock emoji to a role\n` +
                        `\`stats\` : VCs for booster stats.\n`+
                        `\`vars\` : Variables for messages.\n`, inline : true},
                    { name: 'Booster Commands', value: `__For Server Boosters__\n\n`+
                        `\`role\` : Maintain personal role.\n` +
                        `\`lvlrole\` : Role on number of boosts.\n` +
                        `\`rolelist\` : Lists all personal roles.\n`, inline : true},
                    { name: "\u200B", value: "\u200B" },
                    { name: 'User Commands', value: `__For Server Members__\n\n`+
                        `\`help\` : List of Commands.\n` +
                        `\`guide\` : Get started with bot.\n` +
                        `\`count\` : Total server boosts.\n` +
                        `\`invite\` : Bot Invite Links.\n` +
                        `\`userinfo\` : User's Boost Info.\n` +
                        `\`reminder\` : Upvote reminder.\n`, inline : true},
                    { name: "Other Bots" , value: `_ _\n> Check other Multi-Purpose Utility bot that has something for everyone ex - Movies, Games, Events, Music and much more! \n `+
                        `\nhttps://top.gg/bot/699505785847283785`, inline: true },
                    { name: "\u200B" , value: moreInfo },
                )
                .setTimestamp(1610112116000)
                .setFooter('Made With ❤️ By DiabolusGX', client.user.displayAvatarURL({ format: "png", dynamic: true }));
            return message.channel.send( generalHelpEmbed );
        }

        const cmdAliases = [], cmdUsage = [], cmdCooldown = [], cmdPerms = [], cmdDiscriptipn = [];
        const name = args[0].toLowerCase();
        const command = commands.get(name);
        
        if (!command) {
            return message.reply('that\'s not a valid command!');
        }

        if (command.aliases) cmdAliases.push(`${command.aliases.join(' | ')}`);
        if (command.description) cmdDiscriptipn.push(`${command.description}`);
        cmdUsage.push(`\`${dbPrefix} ${command.name} ${command.usage||""}\``);

        cmdCooldown.push(`${command.cooldown || 3} seconds`);

        let title = command.name + ' | ' + cmdAliases;

        let cmdHelpEmbed = new MessageEmbed()
            .setColor("#87339c")
            //.setAuthor(`${message.author.username}`, `${message.author.displayAvatarURL({ format: "png", dynamic: true })}`)
            .setTitle(title)
            .setDescription(`\n${cmdDiscriptipn}\n\n **Usage :** \`${cmdUsage}\`\n ${cmdPerms}`)
            .setThumbnail(client.user.displayAvatarURL({ format: "png", dynamic: true }))
            .addFields(
                { name: "• Cooldown", value : `${cmdCooldown || "`NO COOLDOWN`"}`, inline : true},
                { name: "• Support Server", value: '  https://discord.gg/8kdx63YsDf', inline: true },
                //{ name: "• JSL Bot", value: '[A bot that has something for everyone](https://top.gg/bot/699505785847283785)', inline: true },
            )

        return message.channel.send(cmdHelpEmbed);
    },
};