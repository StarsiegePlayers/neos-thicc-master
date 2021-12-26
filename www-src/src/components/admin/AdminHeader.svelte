
<script>
    export let login;

    const logout = () => {

    }
</script>

<form on:submit|preventDefault={logout}>
    <div class="d-flex flex-column border border-dark border-2 rounded-3 p-3">
        <div class="d-flex flex-row justify-content-center">
            <div class="flex-grow-1">Welcome {$login.username}!</div>
            <div>
                <button class="w-100 btn btn-sm btn-dark px-4" type="submit">Logout</button>
            </div>
        </div>
        <div class="d-flex flex-row justify-content-center">
            <span class="flex-grow-1">{$login.version}</span>
            <div>
                <span id="centerfix" class="px-4">&nbsp;</span>
            </div>
        </div>
    </div>
</form>

<style>
    #centerfix {
        margin: 1.2rem;
    }
</style>