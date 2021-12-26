<script>
    export let settings, form

    const addSTUNServer = (e) => {
        if (e.charCode !== undefined && e.charCode !== 13) {
            return true
        }
        if ($form.stunserver !== "") {
            const stunservers = $settings.Advanced.Network.StunServers
            stunservers.push($form.stunserver)
            $settings.Advanced.Network.StunServers = stunservers
            $form.stunserver = ""
        }
    }

    const removeSTUNServer = (e) => {
        const i = e.target.attributes.index.value
        $form.stunserver = $settings.Advanced.Network.StunServers[i]
        $settings.Advanced.Network.StunServers = $settings.Advanced.Network.StunServers.filter((element, index) => index != i)
    }
</script>

<div class="my-4 text-start">
    <fieldset class="p-2 mt-3 border border-danger border-2 rounded-3 text-start">
        <legend>Advanced Settings</legend>
        <div class="form-check form-switch">
            <input class="form-check-input" type="checkbox" id="advanced.enabled" bind:checked={$form.advanced} />
            <label class="form-check-label" for="advanced.enabled">Edit Advanced Settings</label>
        </div>
        {#if $form.advanced}
            <div class="form-check form-switch">
                <input class="form-check-input" type="checkbox" id="advanced.verbose" bind:checked={$settings.Advanced.Verbose} />
                <label class="form-check-label" for="advanced.verbose">Enable Verbose Debugging Messages</label>
            </div>
            <fieldset class="p-2 mt-3 border border-secondary border-2 rounded-3 text-start">
                <legend class="h5">Network</legend>
                <label for="advanced.network.connectiontimeout" class="form-label text-start">Connection Timeout</label>
                <input id="advanced.network.connectiontimeout" class="form-control" placeholder="2s" bind:value={$settings.Advanced.Network.ConnectionTimeout}>
                <label for="advanced.network.maxpacketsize" class="form-label text-start">Max Packet Size (in bytes)</label>
                <input id="advanced.network.maxpacketsize" class="form-control" placeholder="512" bind:value={$settings.Advanced.Network.MaxPacketSize}>
                <label for="advanced.network.maxbuffersize" class="form-label text-start">Max Buffer Size (in bytes)</label>
                <input id="advanced.network.maxbuffersize" class="form-control" placeholder="32768" bind:value={$settings.Advanced.Network.MaxBufferSize}>
            </fieldset>
            <fieldset class="p-2 mt-3 border border-2 rounded-3 border-secondary text-start">
                <legend class="h5">STUN Servers</legend>
                <div class="d-flex flex-column justify-content-center">
                    {#each $settings.Advanced.Network.StunServers as server, index}
                        <div class="d-flex flex-row justify-content-center">
                            <div class="flex-grow-1">
                                <input index="{index}" type="text" readonly class="form-control-plaintext" value={server}>
                            </div>
                            <div>
                                <button index="{index}" class="btn btn-sm btn-outline-danger" on:click|preventDefault={removeSTUNServer} type="button">-</button>
                            </div>
                        </div>
                    {/each}
                </div>
                <div class="d-flex flex-row justify-content-center">
                    <div class="flex-grow-1">
                        <input id="STUNServerInput" class="form-control" placeholder="stun.l.google.com:19302" on:keypress={addSTUNServer} bind:value={$form.stunserver}>
                    </div>
                    <div>
                        <button class="btn btn-sm btn-outline-success" on:click|preventDefault={addSTUNServer} type="button">+</button>
                    </div>
                </div>
            </fieldset>
            <label for="advanced.maintenance.interval" class="form-label text-start">Maintenance Interval</label>
            <input id="advanced.maintenance.interval" class="form-control" placeholder="1m" bind:value={$settings.Advanced.Maintenance.Interval}>
        {/if}
    </fieldset>
</div>