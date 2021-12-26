<script>
    export let settings, form

    const addNetwork = (e) => {
        if (e.charCode !== undefined && e.charCode !== 13) {
            return true
        }
        if ($form.network !== "") {
            const netSubnet = $form.network.split("/")
            if (netSubnet.length !== 2) {
                return false
            }
            const ipBlocks = netSubnet[0].split(".")
            if (ipBlocks.length !== 4) {
                return false
            }
            const networks = $settings.Service.Banned.Networks
            networks.push($form.network)
            $settings.Service.Banned.Networks = networks
            $form.network = ""
        }
    }

    const removeNetwork = (e) => {
        const i = e.target.attributes.index.value
        $form.network = $settings.Service.Banned.Networks[i]
        $settings.Service.Banned.Networks = $settings.Service.Banned.Networks.filter((element, index) => index != i)
    }
</script>

<div class="my-4 text-start">
    <fieldset class="p-2 mt-3 border border-2 rounded-3 text-start">
        <legend class="">Service Settings</legend>
        <fieldset class="p-2 mt-3 border border-light border-2 rounded-3">
            <legend class="h5">Listening IP/Port</legend>
            <label for="service.listen.ip" class="form-label text-start">IP</label>
            <input id="service.listen.ip" class="form-control" placeholder="" bind:value={$settings.Service.Listen.IP}>
            <label for="service.listen.port" class="form-label text-start">Port</label>
            <input id="service.listen.port" class="form-control" placeholder="29000" bind:value={$settings.Service.Listen.Port}>
        </fieldset>
        <fieldset class="p-2 mt-3 border border-light border-2 rounded-3">
            <legend class="h5">Message Templates</legend>
            <label for="service.templates.motd" class="form-label text-start">MOTD</label>
            <textarea id="service.templates.motd" class="form-control" placeholder="" bind:value={$settings.Service.Templates.MOTD}></textarea>
            <label for="service.templates.timeformat" class="form-label text-start">Time Format</label>
            <textarea id="service.templates.timeformat" class="form-control" placeholder="" bind:value={$settings.Service.Templates.TimeFormat}></textarea>
        </fieldset>
        <fieldset class="p-2 mt-3 border border-light border-2 rounded-3">
            <legend class="h5">Customization</legend>
            <label for="service.serverttl" class="form-label text-start">TTL</label>
            <input id="service.serverttl" class="form-control" placeholder="5m" bind:value={$settings.Service.ServerTTL} />
            <label for="service.id" class="form-label text-start">ID</label>
            <input id="service.id" class="form-control" placeholder="69" bind:value={$settings.Service.ID} />
            <label for="service.serversperip" class="form-label text-start">ID</label>
            <input id="service.serversperip" class="form-control" placeholder="15" bind:value={$settings.Service.ServersPerIP} />
        </fieldset>
        <fieldset class="p-2 mt-3 border border-light border-2 rounded-3">
            <legend class="h5">Banned User Options</legend>
            <label for="service.banned.message" class="form-label text-start">Display Message</label>
            <textarea id="service.banned.message" class="form-control" placeholder="" bind:value={$settings.Service.Banned.Message}></textarea>
            <fieldset class="p-2 mt-3 border border-2 rounded-3 border-secondary text-start">
                <legend class="h6">Banned Networks</legend>
                <div class="d-flex flex-column justify-content-center">
                    {#each $settings.Service.Banned.Networks as network, index}
                        <div id="banned.networks.display" class="d-flex flex-row justify-content-center">
                            <div class="flex-grow-1">
                                <input index="{index}" type="text" readonly class="form-control-plaintext" value={network}>
                            </div>
                            <div>
                                <button index="{index}" class="btn btn-sm btn-outline-danger" on:click|preventDefault={removeNetwork} type="button">-</button>
                            </div>
                        </div>
                    {/each}
                </div>
                <div class="d-flex flex-row justify-content-center">
                    <div class="flex-grow-1">
                        <input id="form.service.banned.network.add" class="form-control" placeholder="127.0.0.1/32" on:keypress={addNetwork} bind:value={$form.network}>
                    </div>
                    <div>
                        <button class="btn btn-sm btn-outline-success" on:click|preventDefault={addNetwork} type="button">+</button>
                    </div>
                </div>
            </fieldset>
        </fieldset>
    </fieldset>
</div>
