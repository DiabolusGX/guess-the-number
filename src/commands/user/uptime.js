const { MessageEmbed } = require("discord.js");
const os = require("os");

module.exports = {
	name: "uptime",
	description: "Shows bot's basic stats",
	category: "user",
	guildOnly: true,

	async run(client, message, args, guildConfigDoc) {
		const dbPrefix = guildConfigDoc.prefix;
		const getCPUInfo = () => {
			let user = 0,
				nice = 0,
				sys = 0,
				idle = 0,
				irq = 0,
				total = 0;
			for (let cpu in cpus) {
				if (!cpus.hasOwnProperty(cpu)) continue;
				user += cpus[cpu].times.user;
				nice += cpus[cpu].times.nice;
				sys += cpus[cpu].times.sys;
				irq += cpus[cpu].times.irq;
				idle += cpus[cpu].times.idle;
			}
			total = user + nice + sys + idle + irq;
			return {
				idle,
				total,
				cpus: cpus.length,
			};
		};
		const getCPUUsage = () => {
			const stats1 = getCPUInfo();
			const startIdle = stats1.idle;
			const startTotal = stats1.total;
			return startIdle / startTotal;
		};

		const cpus = os.cpus();
		const cpuFree = Math.round(getCPUUsage() * 100);
		const cpuUsed = 100 - cpuFree;

		const totalMem = Math.round(os.totalmem() / 1024 / 1024);
		const freeMem = Math.round(os.freemem() / 1024 / 1024);
		const usedMem = Math.round(
			process.memoryUsage().heapUsed / (1024 * 1024)
		);

		const responseTime = Date.now() - message.createdTimestamp;

		const statsEmbed = new MessageEmbed()
			.setColor(guildConfigDoc.color || client.colors.pink)
			.setAuthor({
				name: message.member.nickname || message.author.username,
				iconURL: message.author.displayAvatarURL({
					format: "png",
					dynamic: true,
				}),
			})
			.setDescription(
				`Serving in **${
					client.guilds.cache.size
				} guilds** for ${duration(
					client.uptime
				)} since <t:1610170849:D>\n\n` +
					`• CPU: ${cpuUsed}% used with ${cpus.length} cores!\n` +
					`• Memory: ${usedMem}/${totalMem}MB used and ${freeMem}MB free.\n` +
					`• Response time: ${responseTime}ms.\n`
			)
			.setFooter({
				text: `Checkout "${dbPrefix} premium" and top voter will get free premim`,
			});

		return message.channel.send({ embeds: [statsEmbed] });

		function duration(ms) {
			const sec = Math.floor((ms / 1000) % 60).toString();
			const min = Math.floor((ms / (1000 * 60)) % 60).toString();
			const hrs = Math.floor((ms / (1000 * 60 * 60)) % 60).toString();
			const days = Math.floor(
				(ms / (1000 * 60 * 60 * 24)) % 60
			).toString();
			let res = "";
			if (days > 0) res += `${days.padStart(1, "0")} days, `;
			res += `${hrs.padStart(2, "0")}h:${min.padStart(
				2,
				"0"
			)}m:${sec.padStart(2, "0")}s`;
			return res;
		}
	},
};
