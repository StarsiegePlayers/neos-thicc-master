<script>
    import http from "../store/http";
    import {writable} from "svelte/store";
    import Login from "../components/admin/AdminLogin.svelte"
    import Header from "../components/admin/AdminHeader.svelte"
    import ServiceSettings from "../components/admin/ServiceSettings.svelte";
    import LoggingSettings from "../components/admin/LoggingSettings.svelte";
    import PollingSettings from "../components/admin/PollingSettings.svelte";
    import HTTPDSettings from "../components/admin/HTTPDSettings.svelte";
    import AdvancedSettings from "../components/admin/AdvancedSettings.svelte";

    const login = http({
        loggedIn: true,
        error: "",
        username: "admin",
        password: "",
        version: "Version 0.0.1-dirty",
        expiry: "2021-12-25T23:59:59",
    })

    const settings = http({
        "Log": {
            "ConsoleColors": true,
            "File": "mstrsvr.log",
            "Components": ["startup", "shutdown", "heartbeat", "new-server", "maintenance", "daily-maintenance", "httpd", "poll"]
        },
        "Service": {
            "Listen": {
                "IP": "",
                "Port": 29000
            },
            "Hostname": "Neo's DummyThicc Master",
            "Templates": {
                "MOTD": "Welcome to a Testing server for Neo's Dummythiccness{{.NL}}You are the {{.UserNum}} user today.{{.NL}}Current local server time is: {{.Time}}",
                "TimeFormat": "Y-m-d H:i:s T"
            },
            "ServerTTL": "5m",
            "ID": 69,
            "ServersPerIP": 15,
            "Banned": {
                "Networks": ["224.0.0.0/4"],
                "Message": "You've been banned!"
            }
        },
        "Poll": {
            "Enabled": true,
            "Interval": "5m",
            "KnownMasters": ["master1.starsiegeplayers.com:29000", "master2.starsiegeplayers.com:29000", "master3.starsiegeplayers.com:29000", "starsiege1.no-ip.org:29000", "starsiege.noip.us:29000", "southerjustice.dyndns-server.com:29000", "dustersteve.ddns.net:29000", "starsiege.from-tx.com:29000"]
        },
        "HTTPD": {
            "Enabled": true,
            "Listen": {
                "IP": "",
                "Port": ""
            },
            "Admins": {
                Neo: "akjsldfhkljy7uoi23yhui84oy798$"
            },
            "Secrets": {
                "Authentication": "VS9KTm4rX3QrPzg8YEZLVjN9QSNcUW8oeT47VSB3cWlicDdTTkNqZUxSJTIgYmYsL0EzN1MzYF8mWUVDbT91VQ",
                "Refresh": "ZHNta09aYTN8XHVsQU07RiQqYV1VOy46bTx9bU8ifnske1xfWVE9e3JYJEw9KjZGYWtqKz8tZX4sbiAmbyxadQ"
            }
        },
        "Advanced": {
            "Verbose": false,
            "Network": {
                "ConnectionTimeout": "2s",
                "MaxPacketSize": 512,
                "MaxBufferSize": 32768,
                "StunServers": ["stun.l.google.com:19302", "stun1.l.google.com:19302", "stun2.l.google.com:19302", "stun3.l.google.com:19302", "stun4.l.google.com:19302"]
            },
            "Maintenance": {
                "Interval": "1m"
            }
        },
    })

    const form = writable({
        advanced: false,
        masters: "",
        network: "",
        admin: "",
        stunserver: "",
    })

    const adminFormProcess = () => {

    }

</script>

{#if $login.loggedIn}
    <Login login={login} />
{:else}
    <div class="admin-panel bg-primary bg-opacity-50 rounded-3 boarder-3 p-5">
        <Header login={login} />
        <form on:submit|preventDefault={adminFormProcess}>
            <ServiceSettings settings={settings} form={form} />
            <LoggingSettings settings={settings} form={form} />
            <PollingSettings settings={settings} form={form} />
            <HTTPDSettings settings={settings} form={form} />
            <AdvancedSettings settings={settings} form={form} />
            <input class="btn-lg btn-success" type="submit" value="Save Changes">
        </form>
    </div>
{/if}

<style>
    .admin-panel {
        width: 100%;
        max-width: 50vw;
        padding: 15px;
        margin: auto;
        text-align: center;
    }
</style>
