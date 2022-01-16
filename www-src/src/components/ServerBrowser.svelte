<script>
    import http from "../store/http";
    import {onInterval} from "./interval.js";
    import {timeago} from "./timeago"
    const info = http({
        "RequestTime": "",
        "Masters": [],
        "Games": [],
        "Errors": [],
    });

    const masterInfo = http({
        Hostname: "",
        MOTD: "",
        ID: -1,
        Uptime: "",
    })

    const parseMOTD = (motd) => {
        return motd.split("\\n")
    }

    const intervalCallback = () => {
        info.get("/api/v1/multiplayer/servers");
    }

    // pull some initial data
    intervalCallback();
    masterInfo.get("/api/v1/master/info")

    // update every minute
    onInterval(intervalCallback, 1000*60);

    // export our dynamic info up
    export const ServerInfo = info;
</script>

{#if $masterInfo.Hostname !== ""}
    <div class="d-flex flex-column p-lg-3 mt-5 border border-1 rounded-3 header bg-primary bg-opacity-50">
        <div class="d-flex flex-row justify-content-center">
            <h4>This Server</h4>
        </div>
        <div class="d-flex flex-row justify-content-between">
            <div>
                <h3>{$masterInfo.Hostname}</h3>
            </div>
            <div class="align-self-end">
                <small>Hostname</small>
            </div>
        </div>
        <hr>
        <div class="d-flex flex-row justify-content-between">
            <div class="">
                {#each parseMOTD($masterInfo.MOTD) as value}
                    <h6>{value}</h6>
                {/each}
            </div>
            <div class="align-self-end">
                <small>MOTD</small>
            </div>
        </div>
        <hr>
        <div class="d-flex flex-row justify-content-between">
            <div class="">
                <small use:timeago datetime="{$masterInfo.Uptime}" locale="en_US"></small>
            </div>
            <div class="align-self-end">
                <small>Uptime</small>
            </div>
        </div>
    </div>
{/if}

{#if $info.Masters.length <= 0 && $info.Games.length <= 0 && $info.Errors.Length <= 0}
    <div class="row">
        <h2>Server Information Unavailable</h2>
    </div>
{/if}

{#if $info.Masters.length > 0}
    <div class="row table-responsive pt-4">
        <table class="table table-ss-blue table-bordered table-striped table-hover caption-top">
            <caption class="h4">Peer Master Servers</caption>
            <tr class="table-ss-yellow">
                <th scope="col">No.</th>
                <th scope="col">Hostname</th>
                <th scope="col">Server Name</th>
                <th scope="col">MOTD</th>
                <th scope="col">Reported Games</th>
                <th scope="col">Ping</th>
            </tr>
            {#each $info.Masters as master, i}
                <tr>
                    <th scope="row">{i+1}</th>
                    <td>{master.Address}</td>
                    <td>{master.CommonName}</td>
                    <td>{master.MOTD}</td>
                    <td>{Object.keys(master.Servers).length}</td>
                    <td>{Math.floor(master.Ping / 1000000)} ms</td>
                </tr>
            {/each}
        </table>
    </div>
{/if}

{#if $info.Games.length > 0}
    <hr />
    <div class="row table-responsive">
        <table class="table table-ss-blue table-bordered table-striped table-hover caption-top">
            <caption class="h4">Reporting Game Servers</caption>
            <tr class="table-ss-yellow">
                <th scope="col">No.</th>
                <th scope="col">Server Name</th>
                <th scope="col">Started</th>
                <th scope="col">Legacy Clients</th>
                <th scope="col">Players</th>
                <th scope="col">Ping</th>
                <th scope="col">Server Address</th>
            </tr>
            {#each $info.Games as game, i}
                <tr>
                    <th scope="row">{i+1}</th>
                    <td>
                        {#if game.GameStatus.Protected}<span class="sb-protected"></span>{/if}
                        {#if game.GameStatus.Dedicated}<span class="sb-dedicated"></span>{/if}
                        {#if game.GameStatus.Dynamix}<span class="sb-dynamix"></span>{/if}
                        {#if game.GameStatus.WON}<span class="sb-won"></span>{/if}
                        {game.Name}
                    </td>
                    <td>{game.GameStatus.Started ? "yes" : "no"}</td>
                    <td>{game.GameStatus.AllowOldClients ? "yes" : "no"}</td>
                    <td>{game.PlayerCount} / {game.MaxPlayers}</td>
                    <td>{Math.floor(game.Ping / 1000000)} ms</td>
                    <td><a href="starsiege://{game.Address}">{game.Address}</a></td>
                </tr>
            {/each}
        </table>
    </div>
{/if}
{#if $info.Errors.length > 0}
    <hr />
    <div class="row table-responsive">
        <table class="table table-ss-blue table-bordered table-striped table-hover caption-top">
            <caption class="h4">Errors Encountered</caption>
            {#each $info.Errors as error, i}
                <tr>
                    <th scope="row">{i+1}</th>
                    <td>{error}</td>
                </tr>
            {/each}
        </table>
    </div>
{/if}

<style lang="scss">
    @import "../styles/app/variables";
    .header {
      max-width: 100%;
    }

    .row {
        margin-bottom: 1.5rem;
    }

    .table>caption {
        color: white;
        text-align: center;
    }

    .table>:not(caption)>*>span {
        padding: 0 0.1rem !important;
    }

    .table>:not(caption)>*>* {
        background-color: transparent;
        box-shadow: none;
    }

    .table>:not(:first-child) {
        border: inherit;
    }
</style>
