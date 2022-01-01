
<script>
    export let login, form, scrollPos;

    let submitDisabled = false
    const logoutProcess = () => {
        $form.modal.status = "Logging out..."
        submitDisabled = $form.modal.processing = true
        scrollPos = 0
        login.delete("/api/v1/admin/login")
            .then(() => {
                submitDisabled = $form.modal.processing = false
            })
    }
</script>

<form on:submit|preventDefault={logoutProcess}>
    <div class="d-flex flex-row justify-content-center">
        <div class="flex-grow-1">Welcome {$login.Username}!</div>
        <div>
            <button class="w-100 btn btn-sm btn-dark px-4" type="submit" disabled='{submitDisabled}'>Logout</button>
        </div>
    </div>
    <div class="d-flex flex-row justify-content-center">
        <span class="flex-grow-1">{$login.Version}</span>
        <div>
            <span id="centerfix" class="px-4">&nbsp;</span>
        </div>
    </div>
</form>

<style>
    #centerfix {
        margin: 1.2rem;
    }
</style>