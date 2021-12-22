function nowMinus(hours = 0, minutes = 0) {
    let d = new Date();
    d.setHours(d.getHours() - hours)
    d.setMinutes(d.getMinutes() - minutes)
    return d.toUTCString();
}

export default {
    Title: "Starsiege Players - Master Server",

    Logo: {
        Text: "Starsiege Players Logo",
        Image: "/static/img/leftlogo.png",
        Link: "/"
    },

    Discord: {
        GuildID: "297873205316681728",
        Image: "https://discordapp.com/api/guilds/297873205316681728/widget.png?style=banner2",
        SmallImage: "https://discordapp.com/api/guilds/297873205316681728/widget.png",
        Invite: "https://discord.gg/KA4N6J8",
        Theme: "dark",
        Text: "Join us on Discord!",
    },
}