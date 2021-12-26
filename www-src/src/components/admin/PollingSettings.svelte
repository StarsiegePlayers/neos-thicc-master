<script>
    export let settings, form

    const addKnownMaster = (e) => {
        if (e.charCode !== undefined && e.charCode !== 13) {
            return true
        }
        if ($form.masters !== "") {
            const ipPort = $form.masters.split(":")
            if (ipPort.length < 2) {
                return false
            }
            const masters = $settings.Poll.KnownMasters
            masters.push($form.masters)
            $settings.Poll.KnownMasters = masters
            $form.masters = ""
        }
    }

    const removeKnownMaster = (e) => {
        const i = e.target.attributes.index.value
        $form.masters = $settings.Poll.KnownMasters[i]
        $settings.Poll.KnownMasters = $settings.Poll.KnownMasters.filter((element, index) => index != i)
    }

</script>

<div class="my-4 text-start">
    <fieldset class="p-2 mt-3 border border-2 rounded-3 text-start">
        <legend class="">Server Polling Settings</legend>
        <div class="form-check form-switch">
            <input class="form-check-input" type="checkbox" id="poll.enabled" bind:checked={$settings.Poll.Enabled} />
            <label class="form-check-label" for="poll.enabled">Enable External Master Server Polling</label>
        </div>
        {#if $settings.Poll.Enabled}
            <label for="poll.interval" class="form-label text-start">Polling Interval</label>
            <input id="poll.interval" class="form-control" placeholder="5m" bind:value={$settings.Poll.Interval} />
            <fieldset class="p-2 mt-3 border border-2 rounded-3 border-secondary text-start">
                <legend class="h6">Known Masters</legend>
                <div class="d-flex flex-column justify-content-center">
                    {#each $settings.Poll.KnownMasters as master, index}
                        <div class="d-flex flex-row justify-content-center">
                            <div class="flex-grow-1">
                                <input index="{index}" type="text" readonly class="form-control-plaintext" value={master}>
                            </div>
                            <div>
                                <button index="{index}" class="btn btn-sm btn-outline-danger" on:click|preventDefault={removeKnownMaster} type="button">-</button>
                            </div>
                        </div>
                    {/each}
                </div>
                <div class="d-flex flex-row justify-content-center">
                    <div class="flex-grow-1">
                        <input class="form-control" placeholder="master1.starsiegeplayers.com" on:keypress={addKnownMaster} bind:value={$form.masters}>
                    </div>
                    <div>
                        <button class="btn btn-sm btn-outline-success" on:click|preventDefault={addKnownMaster} type="button">+</button>
                    </div>
                </div>
            </fieldset>
        {/if}
    </fieldset>
</div>
