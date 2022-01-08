<script>
    export let settings;
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
            {#each Object.keys($settings.LogList) as key}
                <option value="{$settings.LogList[key].Tag}">{$settings.LogList[key].Tag} - {$settings.LogList[key].Description}</option>
            {/each}
        </select>
    </fieldset>
</div>
