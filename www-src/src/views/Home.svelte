<script>
    import {timeago} from "../components/timeago"
    import ServerBrowser from "../components/ServerBrowser.svelte";
    export let ServerInfo = undefined;
</script>

<div class="row d-flex flex-wrap border-bottom">
    <div class="d-flex flex-row">
        <div>
            <h3>Service Information</h3>
        </div>
        {#if ServerInfo !== undefined && $ServerInfo.RequestTime !== ""}
            <div class="ms-auto align-self-end">Updated: <small use:timeago datetime="{$ServerInfo.RequestTime}" locale="en_US"></small></div>
        {:else}
            <div class="ms-auto align-self-end">Updated: <small>updating...</small></div>
        {/if}
    </div>
</div>

<div class="row">
    <ServerBrowser bind:ServerInfo={ServerInfo} />
</div>