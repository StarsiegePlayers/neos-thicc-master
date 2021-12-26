<script>
    export let settings, form;
</script>

<div class="my-4 text-start">
    <fieldset class="p-2 mt-3 border border-2 rounded-3 text-start">
        <legend class="">Logging Settings</legend>
        <div class="form-check form-switch">
            <input class="form-check-input" type="checkbox" id="log.consolecolors" bind:checked={$settings.Log.ConsoleColors} />
            <label class="form-check-label" for="log.consolecolors">Colorized Console Output</label>
        </div>
        <label for="log.file" class="form-label text-start">Log File Path</label>
        <input id="log.file" class="form-control" placeholder="" bind:value={$settings.Log.File} />
        <label for="log.components" class="form-label text-start">Components to Log</label>
        <select id="log.components" class="form-select mb-3" aria-label="components to log" multiple size="12" bind:value={$settings.Log.Components}>
            <option value="startup">Startup</option>
            <option value="shutdown">Shutdown</option>
            <option value="heartbeat">Server Heartbeats</option>
            <option value="new-server">New Server Registrations</option>
            <option value="maintenance">Maintenance</option>
            <option value="daily-maintenance">Daily Maintenance</option>
            <option value="template">Templates</option>
            <option value="httpd">HTTPD</option>
            <option value="httpd-router">HTTP Routing</option>
            <option value="poll">Poll</option>
            <option value="stun-client">STUN Client</option>
            <option value="default">Miscellaneous</option>
        </select>
    </fieldset>
</div>
