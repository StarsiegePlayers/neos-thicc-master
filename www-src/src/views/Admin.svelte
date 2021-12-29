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
        LoggedIn: false,
        Error: "",
        Username: "",
        Password: "",
        Version: "",
        Expiry: "",
    })
    login.get("/api/v1/admin/login")

    const settings = http({
        "Log": {
            "ConsoleColors": false,
            "File": "",
            "Components": []
        },
        "Service": {
            "Listen": {
                "IP": "",
                "Port": 0
            },
            "Hostname": "",
            "Templates": {
                "MOTD": "",
                "TimeFormat": ""
            },
            "ServerTTL": "",
            "ID": 0,
            "ServersPerIP": 0,
            "Banned": {
                "Networks": [],
                "Message": ""
            }
        },
        "Poll": {
            "Enabled": false,
            "Interval": "",
            "KnownMasters": []
        },
        "HTTPD": {
            "Enabled": false,
            "Listen": {
                "IP": "",
                "Port": ""
            },
            "Admins": {
            },
            "Secrets": {
                "Authentication": "",
                "Refresh": ""
            },
            "MaxRequestsPerMinute": 0,
        },
        "Advanced": {
            "Verbose": false,
            "Network": {
                "ConnectionTimeout": "",
                "MaxPacketSize": 0,
                "MaxBufferSize": 0,
                "StunServers": []
            },
            "Maintenance": {
                "Interval": ""
            }
        },
    })

    const form = writable({
        advanced: false,
        username: "",
        password: "",
        masters: "",
        network: "",
        admin: "",
        stunserver: "",
    })

    const adminFormProcess = () => {
        settings.post("/api/v1/admin/serversettings")
    }

</script>

{#if $login.LoggedIn !== true}
    <Login login={login} form={login} settings={settings} />
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
