
<script>
    export let login;

    const loginProcess = () => {
        login.post("/api/v1/admin/login")
    }

</script>

<div class="form-signin bg-primary bg-opacity-50 rounded-3 boarder-3 p-5">
    <form on:submit|preventDefault={loginProcess}>
        <img class="mb-4" src="/static/img/leftlogo.png" alt="">
        <h1 class="h3 mb-3 fw-normal">Please sign in</h1>

        <div class="form-floating">
            <input bind:value={$login.username} type="username" class="form-control" id="username" placeholder="username" required>
            <label for="username">Username</label>
        </div>
        <div class="form-floating">
            <input bind:value={$login.password} type="password" class="form-control" id="password" placeholder="password" required>
            <label for="password">Password</label>
        </div>
        <div class="checkbox mb-3">
            <label>
                <input type="checkbox" value="remember-me"> Remember me
            </label>
        </div>
        {#if $login.error !== ""}
            <div class="h6 mb-3 fw-normal invalid-text">{$login.error}</div>
        {/if}
        <button class="w-100 btn btn-lg btn-primary" type="submit">Sign in</button>
        <p class="mt-5 mb-3 text-muted">2021 - StarsiegePlayers</p>
    </form>
</div>

<style>
    .form-signin {
        width: 100%;
        max-width: 35vw;
        padding: 15px;
        margin: auto;
        text-align: center;
    }

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