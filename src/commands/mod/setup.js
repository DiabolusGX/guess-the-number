const { MessageEmbed } = require("discord.js");
const guildConfigModel = require("../../database/models/guildConfig");
const find = require("../../utils/find");

module.exports = {

    name: "setup",
    aliases: ["set"],
    description: `This command is used to setup/edit/remove bot settings.\n` +
        `\n**Prefix \`gg setup prefix <new prefix>\`**` +
        `\n\n__Bot Manager__ \`gg setup manager <role | disable>\`\nUsers with this role (and admins) will be able to edit bot settings.\n` +
        `\n__DM__ \`gg setup dm <enable | disable>\`\nBot will DM the winner (Enabled by default).\n` +
        //`\n__Message__ \`gg setup msg <message | disable>\`\nBot will send this message when someone guesses the number.\n` +
        `\n__Win Role__ \`gg setup win-role <role | disable>\`\nThis role will be assigned to the winner of the event.\n` +
        `\n__Required Role__ \`gg setup req-role <role | disable>\`\nUsers only with this role will be eligible to guess the number (If enabled).\n` +
        `\n__Lock Role__ \`gg setup lock-role <role | disable>\`\nBot will lock channel for this role after someone guesses the correct number.\n` +
        `\n**If you want to remove any config, type \`disable\` in value.**\nExample :  \`gg setup win-role disable\` then bot will not assign any role when someone wins.`,
    category: "mod",
    guildOnly: true,
    usage: "< prefix | manager | dm | msg | win-role | req-role | lock-role >",

    async run(client, message, args) {
        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);
        const botRole = message.guild.me.roles.highest;
        const guildDoc = await guildConfigModel.findOne({ id: message.guild.id });

        if (!args[0]) {
            const botManager = guildDoc.botManager ? `<@&${guildDoc.botManager}>` : "**Disabled**";
            const dm = guildDoc.dm ? "**Enabled**" : "**Disabled**";
            //const msg = guildDoc.msg ? guildDoc.msg : "**Disabled**";
            const winRole = guildDoc.winRole ? `<@&${guildDoc.winRole}>` : "**Disabled**";
            const reqRole = guildDoc.reqRole ? `<@&${guildDoc.reqRole}>` : "**Disabled**";
            const lockRole = guildDoc.lockRole ? `<@&${guildDoc.lockRole}>` : "**Disabled**";

            const configEmbed = new MessageEmbed()
                .setColor(client.colors[0])
                .setAuthor({
                    name: `Setup Config for - ${message.guild.name}`,
                    iconURL: client.user.displayAvatarURL({ format: "png", dynamic: true })
                })
                .setDescription(`Current __Prefix__ : **\`${guildDoc.prefix}\`**` +
                    `\n\n**__BOT Manager__** : ${botManager} \n Can edit, remove bot settings` +
                    `\n\n**__DM__** : ${dm} \nWether bot will send DM to the winner or not.` +
                    //`\n\n**__Message__** : ${msg} \n__Message that bot will send when someone guesses the number__.` +
                    `\n\n**__Win Role__** : ${winRole} \nRole that'll be assigned to user who guesses the number.` +
                    `\n\n**__Required Role__** : ${reqRole} \n Bot will only accept guesses from users in this role onlly.` +
                    `\n\n**__Lock Role__** : ${lockRole} \n Bot will lock channel after someone guesses the number for this role.` +
                    `\n\n__For more info - \`${dbPrefix} help setup\`__ \n\n`)
                .setThumbnail(message.guild.iconURL({ format: "png", dynamic: true, size: 1024 }) || client.user.displayAvatarURL({ format: "png", dynamic: true }))
                .setFooter({
                    text: `Made with ❤️ by DiabolusGX`,
                    iconURL: client.user.displayAvatarURL({ format: "png", dynamic: true })
                });

            return message.channel.send({ embeds: [configEmbed] });
        }

        if (args[0] === "prefix") {
            if (!args[1]) return message.channel.send(`${client.myEmojis[1]} | **Please provide the new prefix you want to set.**\n> **\`${dbPrefix} setup prefix !\`** will change prefix to \`!\``);
            if (args[1] === "none" || args[1] === "disable") return message.channel.send(`Prefix can't be \`disable\`.`);

            const currPrefix = client.guildConfigPrefix.get(message.guild.id);
            guildConfigModel.updateOne({ id: message.guild.id }, { $set: { prefix: args[1] } }, (err) => console.error);
            client.guildConfigPrefix.set(message.guild.id, args[1]);

            const prefixUpdateEmbed = new MessageEmbed()
                .setColor(client.colors[0])
                .setAuthor({ name: `Prefix Update` })
                .setDescription(`\nChanged prefix from : \`${currPrefix}\` to : \`${args[1]}\` \n\n >>> Prefix will be required to use commands!`)
            return message.channel.send({ embeds: [prefixUpdateEmbed] });
        }
        else if (args[0] === "manager") {
            if (!args[1]) return message.channel.send(`${client.myEmojis[1]} | **Please mention a role (or use role id) to allow users with that role to edit/remove bot settings**\n` +
                `> **\`${dbPrefix} setup manager @mods\`** will allow \`mods\` role to manage bot settings!`);
            let botManagerRole;
            let argsJoin = args.slice(1).join(" ");
            if (argsJoin === "none" || argsJoin === "disable") guildConfigModel.updateOne({ id: message.guild.id }, { $set: { botManager: null } }, (err) => console.error);
            else {
                botManagerRole = await find.getRole(message, argsJoin);
                if (!botManagerRole) return message.channel.send(`I can't find role with NAME or ID : \`${argsJoin}\``);
                guildConfigModel.updateOne({ id: message.guild.id }, {
                    $set: {
                        botManager: botManagerRole.id
                    }
                }, (err) => console.error);
            }

            message.channel.send({
                content: "__**Note**__ : Make sure you give this role to trusted users only. They can manage bot settings."
            });
            const managerRoleUpdateEmbed = new MessageEmbed()
                .setColor(client.colors[0])
                .setAuthor({ name: `Bot Manager Role Update` })
                .setDescription(`\nChanged Manager Role to : **${botManagerRole || "disable"}** \n\n >>> User with this role can modify bot settings!\n(If \`disable\` then only Admins can edit setting)`)
            return message.channel.send({ embeds: [managerRoleUpdateEmbed] });
        }
        else if (args[0] === "dm") {
            if (!args[1]) return message.channel.send({
                content: `${client.myEmojis[1]} | **Please choose from : \`ENABLE\` or \`DISABLE\` for DM**` +
                    `\n > **\`${dbPrefix} setup dm enable\`** __(default)__ will DM who guesses the number.` +
                    `\n > **\`${dbPrefix} setup dm disable\`** will only send normal message`
            });
            let desc;
            if (args[1] === "enable") {
                guildConfigModel.updateOne({ id: message.guild.id }, { $set: { dm: true } }, (err) => console.error);
                desc = "\n**Enabled** DM!\nA DM will be sent when someone guesses the correct number.";
            }
            else if (args[1] === "disable") {
                guildConfigModel.updateOne({ id: message.guild.id }, { $set: { dm: false } }, (err) => console.error);
                desc = "\n**Disabled** DM!\nOnly a message will be sent when someone guesses the correct number.";
            }
            else return message.channel.send({
                content: `${client.myEmojis[1]} | **Please choose from : \`ENABLE\` or \`DISABLE\` for DM**` +
                    `\n > **\`${dbPrefix} setup dm enable\`** __(default)__ will DM who guesses the number.` +
                    `\n > **\`${dbPrefix} setup dm disable\`** will only send normal message`
            });

            const dmUpdateEmbed = new MessageEmbed()
                .setColor(client.colors[0])
                .setAuthor({ name: `Updated DM` })
                .setDescription(desc);
            return message.channel.send({ embeds: [dmUpdateEmbed] });
        }
        // else if (args[0] === "msg" || args[0] === "message") {
        //     if (!args[1]) return message.channel.send(`**Please enter the message you want to be sent when someone guesses the number**` +
        //         `\n > **\`${dbPrefix} setup msg Thanks for participating everyone and Congratulations {user}, You've has guessed the currect number i.e {number}\`** ` +
        //         `\n Bot will send this message when someone guesses the correct number!`);
        //     let desc, msg = args.slice(1).join(" ");
        //     if (args[1] === "none" || args[1] === "disable") {
        //         guildConfigModel.updateOne({ id: message.guild.id }, { $set: { msg: null } }, (err) => console.error);
        //         desc = "Disabled Message!\nEnter new message to enable message on victory.";
        //     }
        //     else {
        //         if (guildDoc.premium && greetDoc.messages.length >= 10) return message.channel.send("<:redtick:803175735665623091> | **Limit Error**" +
        //             "\nYou can only add upto 10 messages!\n Join Support Server to get additional help.\n https://discord.com/invite/8kdx63YsDf");
        //         else if (greetDoc.messages.length >= 3) return message.channel.send("<:redtick:803175735665623091> | **Premium Error**" +
        //             "\nYou need premium to add more then 3 Greet Messages!\n Join Support Server to get premium.\n https://discord.com/invite/8kdx63YsDf");
        //         guildConfigModel.updateOne({ id: message.guild.id }, { $set: { msg: msg } }, (err) => console.error);
        //         desc = `Added message : \n\n>>> ${msg}`;
        //     }

        //     const tip = "**Tip** : Use `" + dbPrefix + " variables` to get list of all variables you can use.";
        //     const messageUpdateEmbed = new MessageEmbed()
        //         .setColor(client.colors[0])
        //         .setAuthor(`Updated Message`)
        //         .setDescription(desc);
        //     return message.channel.send(tip, messageUpdateEmbed);
        // }
        else if (args[0] === "win-role") {
            if (!args[1]) return message.channel.send({
                content: `${client.myEmojis[1]} | **Please mention (or role id) the role which you want to give to winners.**\n` +
                    `> **\`${dbPrefix} setup win-role @guess-winner\`** Bot will assign this role when someone guesses the correct number.`
            });
            let winRole, desc = "";
            const argsJoin = args.slice(1).join(" ");
            if (argsJoin.toLowerCase() === "disable") {
                guildConfigModel.updateOne({ id: message.guild.id }, { $set: { winRole: null } }, (err) => console.error);
                desc = "Updated Win Role : **Disabled** win role.\ni.e No role will be assigned when someone guess the number.";
            }
            else {
                winRole = await find.getRole(message, argsJoin);
                if (!winRole) return message.channel.send({
                    content: `${client.myEmojis[1]} | I can't find role with NAME or ID : \`${argsJoin}\``
                });
                if (winRole.position >= botRole.position) return message.channel.send({
                    content: client.myEmojis[1] + "| Bot's highest role is not high enough to assign **" + winRole.name + "** role.\n" +
                        "Please make sure bot has `MANAGE_ROLES` perms and bot's role is high enough in role hierarchy."
                });
                guildConfigModel.updateOne({ id: message.guild.id }, { $set: { winRole: winRole.id } }, (err) => console.error);
                desc = `Updated Win Role : **${winRole.toString()}** \n\n >>> Bot will assign this role when someone guesses the number`;
            }

            const winRoleUpdateEmbed = new MessageEmbed()
                .setColor(client.colors[0])
                .setAuthor({ name: `WIN Role Update` })
                .setDescription(`\n${desc}`)
            return message.channel.send({ embeds: [winRoleUpdateEmbed] });
        }
        else if (args[0] === "req-role") {
            if (!args[1]) return message.channel.send({
                content: `${client.myEmojis[1]} | **Please mention (or role id) the role which is required to guess the number.**\n` +
                    `> **\`${dbPrefix} setup req-role @guessers\`** Bot will accept guesses only from users in this role.`
            });
            let reqRole, desc = "";
            const argsJoin = args.slice(1).join(" ");
            if (argsJoin.toLowerCase() === "disable") {
                guildConfigModel.updateOne({ id: message.guild.id }, { $set: { reqRole: null } }, (err) => console.error);
                desc = "Updated Required Role : **Disabled** required role.\ni.e everyone can guess the number.";
            }
            else {
                reqRole = await find.getRole(message, argsJoin);
                guildConfigModel.updateOne({ id: message.guild.id }, { $set: { reqRole: reqRole.id } }, (err) => console.error);
                desc = `Updated Required Role : **${reqRole.toString()}** \n\n >>> Bot will accept guesses from users in this role only.`;
            }

            const reqRoleUpdateEmbed = new MessageEmbed()
                .setColor(client.colors[0])
                .setAuthor({ name: `Required Role Update` })
                .setDescription(`\n${desc}`)
            return message.channel.send({ embeds: [reqRoleUpdateEmbed] });
        }
        else if (args[0] === "lock-role") {
            if (!args[1]) return message.channel.send({
                content: `${client.myEmojis[1]} | **Please mention (or role id) the role that bot will lock channel for.**\n` +
                    `> **\`${dbPrefix} setup req-role @everyone\`** Bot will lock channel for everyone after someone guesses the number.`
            });
            let lockRole, desc = "";
            const argsJoin = args.slice(1).join(" ");
            if (argsJoin.toLowerCase() === "disable") {
                guildConfigModel.updateOne({ id: message.guild.id }, { $set: { lockRole: null } }, (err) => console.error);
                desc = "Updated Lock Role : **Disabled** required role.\ni.e Channel will not be locked.";
            }
            else {
                if (argsJoin.toLowerCase() === "everyone") lockRole = message.guild.roles.everyone;
                else lockRole = await find.getRole(message, argsJoin);
                if (!lockRole) return message.channel.send({
                    content: `${client.myEmojis[1]} | **No Role Found** matching with \`${argsJoin}\``
                });
                guildConfigModel.updateOne({ id: message.guild.id }, { $set: { lockRole: lockRole.id } }, (err) => console.error);
                desc = `Updated Lock Role : **${lockRole.toString()}** \n\n >>> Bot will lock channel for this given role.`;
            }

            const reqRoleUpdateEmbed = new MessageEmbed()
                .setColor(client.colors[0])
                .setAuthor({ name: `LOCK Role Update` })
                .setDescription(`\n${desc}`)
            return message.channel.send({ embeds: [reqRoleUpdateEmbed] });
        }

        else return message.channel.send({
            content: client.myEmojis[1] + " | **No vaid option.**\nCheck options from command usage and try again!\n> `" + dbPrefix + " setup < prefix | manager | dm | msg | win-role | req-role >`."
        });
    }
}