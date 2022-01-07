<script>
    export let login, form, settings, scrollPos;

    let submitDisabled = false
    const loginProcess = () => {
        $form.modal.status = "Logging In..."
        submitDisabled = $form.modal.processing = true
        scrollPos = 0
        login.post("/api/v1/admin/login", {
                "Username": $form.username,
                "Password": $form.password,
            })
            .then(() => {
                if ($login.LoggedIn) {
                    $form.modal.status = "Fetching server settings..."
                    settings.get("/api/v1/admin/serversettings")
                        .then(() => {
                            submitDisabled = $form.modal.processing = false
                        })
                    return
                }
                submitDisabled = $form.modal.processing = false
            })
        $form.password = ""
    }
</script>

<form on:submit|preventDefault={loginProcess}>
    <img class="mb-4" src="/static/img/leftlogo.png" alt="">
    <h1 class="h3 mb-3 fw-normal">Please sign in</h1>

    <div class="form-floating">
        <input bind:value={$form.username} type="username" class="form-control" id="username" placeholder="username" required>
        <label for="username">Username</label>
    </div>
    <div class="form-floating">
        <input bind:value={$form.password} type="password" class="form-control" id="password" placeholder="password" required>
        <label for="password">Password</label>
    </div>
    <div class="checkbox mb-3">
        <label>
            <input type="checkbox" value="remember-me"> Remember me
        </label>
    </div>
    {#if $login.Error !== ""}
        <div class="h6 mb-3 fw-normal invalid-text">{$login.Error}</div>
    {/if}
    <button class="w-100 btn btn-lg btn-primary" type="submit" disabled={submitDisabled}>Sign in</button>
    <p class="mt-5 mb-3 text-muted">2021 - StarsiegePlayers</p>
</form>

<style>
    .invalid-text {
        color: red;
    }

    .checkbox {
        font-weight: 400;
        text-align: left;
    }

    .form-floating:focus-within {
        z-index: 2;
    }

    input[type="username"] {
        margin-bottom: -1px;
        border-bottom-right-radius: 0;
        border-bottom-left-radius: 0;
    }

    input[type="password"] {
        margin-bottom: 10px;
        border-top-left-radius: 0;
        border-top-right-radius: 0;
    }
</style>