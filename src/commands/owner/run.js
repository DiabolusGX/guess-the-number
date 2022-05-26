const axios = require("axios");

module.exports = {

    name: "run",
    aliases: [],
    description: "Evaluate code / any equation.",
    category: "owner",
    guildOnly: true,
    args: true,
    usage: "<code>",

    async run(client, message, args) {

        if (message.author.id !== "454611998051794954") return;

        const clean = (text) => {
            if (typeof text === "string") return text.replace(/`/g, "`" + String.fromCharCode(8203)).replace(/@/g, "@" + String.fromCharCode(8203));
            else return text;
        };

        try {
            const code = args.join(" ");
            let evaled = await eval(code);

            if (typeof evaled !== "string") evaled = require("util").inspect(evaled);
            if (evaled.includes(process.env.BOT_TOKEN)) return message.channel.send(`\`\`\`js\n${"NDM4NDM4MzgzMzY3MzU2NDE2.DjEOFg.Qhu4upnvn_C5feYz2wsdN09QKUI"}\`\`\``);

            message.channel.send(clean(evaled), { code: "js" });
        } catch (err) {
            message.channel.send(`\`ERROR\` \`\`\`js\n${clean(err)}\n\`\`\``);
        }
    },
};