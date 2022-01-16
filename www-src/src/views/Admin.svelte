<script>
    import http from "../store/http";
    import {writable} from "svelte/store";
    import { fade } from "svelte/transition";

    import { Rainbow } from 'svelte-loading-spinners'
    import Modal from "../components/Modal.svelte";
    import Login from "../components/admin/AdminLogin.svelte"
    import Header from "../components/admin/AdminHeader.svelte"
    import ServiceSettings from "../components/admin/ServiceSettings.svelte";
    import LogSettings from "../components/admin/LogSettings.svelte";
    import PollSettings from "../components/admin/PollSettings.svelte";
    import HTTPDSettings from "../components/admin/HTTPDSettings.svelte";
    import AdvancedSettings from "../components/admin/AdvancedSettings.svelte";
    import HeaderMessage from "../components/admin/HeaderMessage.svelte";

    const login = http({
        "LoggedIn": false,
        "Username": "",
        "Password": "",
        "Version": "",
        "Expiry": "",
        "Error": "",
        "ErrorCode": 0,
    })
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
        "LogList":{},
        "Error": "",
        "ErrorCode": 0,
    })
    const powerActions = http({
        "Action": "",
        "Error": "",
        "ErrorCode": 0,
    })
    const form = writable({
        modal: {
            processing: false,
            status: "Processing...",
            scrollPos: 0,
        },
        powerActionModal: {
            show: false,
            confirm: "",
        },
        advanced: false,
        username: "",
        password: "",
        masters: "",
        network: "",
        admin: "",
        stunserver: "",
    })

    let submitDisabled = false, displayResult = false, resultSuccess = true, scrollPos
    login.get("/api/v1/admin/login")
        .then(() => {
            if ($login.LoggedIn) {
                $form.modal.status = "Fetching server settings..."
                scrollPos = 0
                submitDisabled = $form.modal.processing = true
                settings.get("/api/v1/admin/serversettings")
                    .then(() => {
                        submitDisabled = $form.modal.processing = false
                    })
            }
        })

    const adminFormProcess = () => {
        scrollPos = 0
        submitDisabled = $form.processing = true
        settings.post("/api/v1/admin/serversettings", $settings)
            .then(() => {
                if (($settings.Error !== undefined && $settings.Error !== "") || ($settings.ErrorCode !== undefined && $settings.ErrorCode !== 0)) {
                    resultSuccess = false
                }
                submitDisabled = $form.processing = false
                displayResult = true
                setTimeout(() => { displayResult = false }, 3000)
            })
    }

    const adminFormPowerAction = (e) => {
        scrollPos = 0
        submitDisabled = $form.powerActionModal.show = true
        $powerActions.Action = e.submitter.value.toLowerCase()
    }

    const adminFormPowerActionSubmit = () => {
        powerActions.post("/api/v1/admin/poweraction", $powerActions)
            .then(() => {
                submitDisabled = $form.powerActionModal.show = false
                displayResult = true
                setTimeout(() => { displayResult = false }, 3000)
            })
    }
</script>

<svelte:window bind:scrollY={scrollPos}/>
<Modal bind:open={$form.modal.processing}>
    <div class="modal-header bg-primary">
        <h5 class="modal-title">{$form.modal.status}</h5>
    </div>
    <div class="modal-body d-flex flex-row justify-content-center bg-primary text-center">
        <Rainbow size="60" color="#18303f" />
    </div>
</Modal>

<Modal bind:open={$form.powerActionModal.show}>
    <div class="modal-header bg-primary">
        <h5 class="modal-title">Are you absolutely sure?</h5>
    </div>
    <div class="modal-body d-flex flex-row justify-content-center bg-primary">
        <form on:submit|preventDefault={adminFormPowerActionSubmit}>
            <p class="text-danger p-5 bg-warning bg-opacity-25 rounded rounded-3 text-center">
                You are about to {$powerActions.Action.toUpperCase()} the server {$settings.Service.Hostname.replaceAll("\\n", "")}
            </p>
            <p>
                <label for="admin.poweraction.confirm" class="form-label text-start">Please type {$powerActions.Action.toUpperCase()} to confirm.</label>
                <input id="admin.poweraction.confirm" class="form-control" placeholder="" bind:value={$form.powerActionModal.confirm} />
            </p>
            <div class="d-flex flex-row justify-content-between">
                <input class="text-danger" type="submit" disabled="{$form.powerActionModal.confirm !== $powerActions.Action.toUpperCase()}" value="{$powerActions.Action.toUpperCase()}" />
                <input class="btn-secondary" type="button" value="Cancel" on:click|preventDefault={()=>{submitDisabled = $form.powerActionModal.show = false}}>
            </div>
        </form>
    </div>
</Modal>

{#if $login.LoggedIn !== true}
    <div class="form-signin bg-primary bg-opacity-50 rounded-3 boarder-3 p-5">
        <Login login={login} form={form} settings={settings} scrollPos={scrollPos} />
    </div>
{:else}
    <div class="admin-panel bg-primary bg-opacity-50 rounded-3 boarder-3 p-5" transition:fade>
        <div class="d-flex flex-column border border-dark border-2 rounded-3 p-3">
            <Header login={login} form={form} scrollPos={scrollPos} />
        </div>
        {#if displayResult}
        <HeaderMessage success={resultSuccess}>
            {#if (resultSuccess)}
                Settings saved successfully
            {:else}
                {$settings.Error}
            {/if}
        </HeaderMessage>
        {/if}
        <form on:submit|preventDefault={adminFormProcess}>
            <ServiceSettings settings={settings} form={form} />
            <LogSettings settings={settings} form={form} />
            <PollSettings settings={settings} form={form} />
            <HTTPDSettings settings={settings} form={form} />
            <AdvancedSettings settings={settings} form={form} />
            <input class="btn-lg btn-success" type="submit" value="Save Changes" disabled='{submitDisabled}'>
        </form>
        <form on:submit|preventDefault={adminFormPowerAction}>
            <input class="btn-lg btn-danger" type="submit" value="Shutdown" disabled='{submitDisabled}'>
            <input class="btn-lg btn-warning" type="submit" value="Restart" disabled='{submitDisabled}'>
        </form>
    </div>
{/if}

<style>
    .form-signin {
        z-index: -1;
        width: 100%;
        max-width: 35vw;
        padding: 15px;
        margin: auto;
        text-align: center;
    }

    .admin-panel {
        width: 100%;
        max-width: 50vw;
        padding: 15px;
        margin: auto;
        text-align: center;
    }
</style>
