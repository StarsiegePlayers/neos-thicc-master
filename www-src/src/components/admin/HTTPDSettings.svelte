<script>
    export let settings, form

    const addAdmin = (e) => {
        if (e.charCode !== undefined && e.charCode !== 13) {
            return true
        }
        if ($form.admin !== "") {
            const usrPw = $form.admin.split(":")
            if (usrPw.length < 2) {
                return false
            }
            $settings.HTTPD.Admins[usrPw[0]] = usrPw[1]
            $form.admin = ""
        }
    }

    const removeAdmin = (e) => {
        const i = e.target.attributes.index.value
        const entries = Object.entries($settings.HTTPD.Admins)
        $form.admin = entries[i][0] + ":" + entries[i][1]

        $settings.HTTPD.Admins = Object.fromEntries(Object.entries($settings.HTTPD.Admins).filter((key, index) => index != i))
    }

    const onFocus = (e) => {
        e.target.attributes.type.value = "text"
    }

    const onBlur = (e) => {
        e.target.attributes.type.value = "password"
    }
</script>

<div class="my-4 text-start">
    <fieldset class="p-2 mt-3 border border-2 rounded-3 text-start">
        <legend class="">HTTPD Settings</legend>
        <div class="form-check form-switch">
            <input class="form-check-input" type="checkbox" id="httpd.enabled" bind:checked={$settings.HTTPD.Enabled} />
            <label class="form-check-label" for="httpd.enabled">Enable HTTPD Service</label>
        </div>
        {#if $settings.HTTPD.Enabled}
            <fieldset class="p-2 mt-3 border border-light border-2 rounded-3">
                <legend class="h5">Listening IP/Port</legend>
                <label for="httpd.listen.ip" class="form-label text-start">IP</label>
                <input id="httpd.listen.ip" class="form-control" placeholder="" bind:value={$settings.HTTPD.Listen.IP}>
                <label for="httpd.listen.port" class="form-label text-start">Port</label>
                <input id="httpd.listen.port" class="form-control" placeholder="" bind:value={$settings.HTTPD.Listen.Port}>
            </fieldset>
            <fieldset class="p-2 mt-3 border border-2 rounded-3 border-secondary text-start">
                <legend class="h6">Admin Users</legend>
                <div class="d-flex flex-column justify-content-center">
                    {#each Object.entries($settings.HTTPD.Admins) as [user, hash], index}
                        <div class="d-flex flex-row justify-content-center">
                            <div class="flex-grow-1">
                                <input index="{index}" type="text" readonly class="form-control-plaintext" value={user+":"+hash}>
                            </div>
                            <div>
                                <button index="{index}" class="btn btn-sm btn-outline-danger" on:click|preventDefault={removeAdmin} type="button">-</button>
                            </div>
                        </div>
                    {/each}
                </div>
                <div class="d-flex flex-row justify-content-center">
                    <div class="flex-grow-1">
                        <input id="adminUserInput" class="form-control" placeholder="username:passwordhash" on:keypress={addAdmin} bind:value={$form.admin}>
                    </div>
                    <div>
                        <button class="btn btn-sm btn-outline-success" on:click|preventDefault={addAdmin} type="button">+</button>
                    </div>
                </div>
            </fieldset>
            <fieldset class="p-2 mt-3 border border-light border-2 rounded-3">
                <legend class="h5">Cookie Authentication Secrets</legend>
                <label for="httpd.secrets.authentication">Primary Secret</label>
                <input bind:value={$settings.HTTPD.Secrets.Authentication} on:focus={onFocus} on:blur={onBlur} type="password" class="form-control" id="httpd.secrets.authentication" placeholder="">
                <label for="httpd.secrets.refresh">Refresh Secret</label>
                <input bind:value={$settings.HTTPD.Secrets.Refresh} on:focus={onFocus} on:blur={onBlur} type="password" class="form-control" id="httpd.secrets.refresh" placeholder="">
            </fieldset>
            <label for="httpd.maxrequestsperminute" class="form-label text-start">Max Requests Per Minute Per IP</label>
            <input id="httpd.maxrequestsperminute" class="form-control" placeholder="15" bind:value={$settings.HTTPD.MaxRequestsPerMinute}>
        {/if}
    </fieldset>
</div>
