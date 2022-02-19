const { MessageEmbed } = require("discord.js");

module.exports = {

    name: "help",
    aliases: ["commands"],
    description: "Shows all available commands.",
    category: "user",
    guildOnly: false,

    async run(client, message, args) {

        const dbPrefix = client.guildConfigPrefix.get(message.guild.id);

        const discription = []; const moreInfo = []; const allCommandNames = [];
        const { commands } = message.client;

        if (!args.length) {
            for (let key of commands.keys()) allCommandNames.push(key);
            allCommandNames.join("` `");
            moreInfo.push(`\nYou can send \`${dbPrefix} help [command name]\` to get info on a specific command!`);
            discription.push(`\nTo get started with stup - **__\`${dbPrefix} help setup\`__**\n\n`);

            const generalHelpEmbed = new MessageEmbed()
                .setColor(client.colors[0])
                .setDescription(`${discription}\n\n`)
                .addFields(
                    { name: "\u200B", value: "_ _" },
                    {
                        name: 'Setup Commands', value: `__For Admins and Bot Manager__\n\n` +
                            `\`setup\` : Setup Server settings.\n` +
                            `\`start\` : To start the game.\n` +
                            `\`hint\` : Gives hint for running game.\n` +
                            `\`end\` : End ongoing game.\n`, inline: true
                    },
                    {
                        name: 'User Commands', value: `__For Server Members__\n\n` +
                            `\`help\` : List of Commands.\n` +
                            `\`game\` : Running game info.\n` +
                            `\`info\` : User games & points info.\n` +
                            `\`lb\` : Leader Board based on points.\n` +
                            `\`invite\` : Bot Invite Links.\n`, inline: true
                    },
                    {
                        name: "Other Bots", value: `> Check **Booster Bot** : \n> A bot that handles everything related to server boosts! \n ` +
                            `🔗  https://boosterbot.xyz/`
                    },
                    { name: "\u200B", value: moreInfo },
                )
                .setTimestamp(1610112116000)
                .setFooter({
                    text : 'Made With ❤️ By DiabolusGX',
                    iconURL: client.user.displayAvatarURL({ format: "png", dynamic: true })
                });
            return message.channel.send({ embeds: [generalHelpEmbed] });
        }

        const cmdAliases = [], cmdUsage = [], cmdCooldown = [], cmdPerms = [], cmdDiscriptipn = [];
        const name = args[0].toLowerCase();
        const command = commands.get(name);

        if (!command) {
            return message.reply('that\'s not a valid command!');
        }

        if (command.aliases) cmdAliases.push(`${command.aliases.join(' | ')}`);
        if (command.description) cmdDiscriptipn.push(`${command.description}`);
        cmdUsage.push(`\`${dbPrefix} ${command.name} ${command.usage || ""}\``);

        cmdCooldown.push(`${command.cooldown || 3} seconds`);

        const title = command.name + ' | ' + cmdAliases;

        const cmdHelpEmbed = new MessageEmbed()
            .setColor(client.colors[0])
            .setTitle(title)
            .setDescription(`\n${cmdDiscriptipn}\n\n **Usage :** \`${cmdUsage}\`\n ${cmdPerms}`)
            .setThumbnail(client.user.displayAvatarURL({ format: "png", dynamic: true }))
            .addFields(
                { name: "• Cooldown", value: `${cmdCooldown || "`NO COOLDOWN`"}`, inline: true },
                { name: "• Support Server", value: '  https://discord.gg/8kdx63YsDf', inline: true },
            )

        return message.channel.send({ embeds: [cmdHelpEmbed] });
    },
};