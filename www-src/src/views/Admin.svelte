<script>
    import http from "../store/http";
    const login = http({
        "RequestTime": "",
        "Errors": [],
        "User": {},
    })

    let state = {
        login: {
            loggedIn: false,
            error: false,
            form: {
                Username: "",
                Password: "",
            },
            process: function() {
                login.post("/api/v1/admin/login", state.login.form)
            }
        },
    }

</script>

{#if !state.login.loggedIn}
    <div class="form-signin bg-primary bg-opacity-50 rounded-3 boarder-3 p-5">
        <form on:submit|preventDefault={state.login.process}>
            <img class="mb-4" src="/static/img/leftlogo.png" alt="">
            <h1 class="h3 mb-3 fw-normal">Please sign in</h1>

            <div class="form-floating">
                <input bind:value={state.login.form.Username} type="email" class="form-control" id="floatingInput" placeholder="user" required>
                <label for="floatingInput">Username</label>
            </div>
            <div class="form-floating">
                <input bind:value={state.login.form.Password} type="password" class="form-control" id="floatingPassword" placeholder="password" required>
                <label for="floatingPassword">Password</label>
            </div>

            <div class="checkbox mb-3">
                <label>
                    <input type="checkbox" value="remember-me"> Remember me
                </label>
            </div>
            {#if $login.errors.Length > 0}
                <div class="h5 mb-3 fw-normal">Invalid Username or Password, please try again</div>
            {/if}
            <button class="w-100 btn btn-lg btn-primary" type="submit">Sign in</button>
            <p class="mt-5 mb-3 text-muted">2021 - StarsiegePlayers</p>
        </form>
    </div>
{/if}

<style>
    .form-signin {
        width: 100%;
        max-width: 35vw;
        padding: 15px;
        margin: auto;
        text-align: center;
    }

    .form-signin .checkbox {
        font-weight: 400;
        text-align: left;
    }

    .form-signin .form-floating:focus-within {
        z-index: 2;
    }

    .form-signin input[type="email"] {
        margin-bottom: -1px;
        border-bottom-right-radius: 0;
        border-bottom-left-radius: 0;
    }

    .form-signin input[type="password"] {
        margin-bottom: 10px;
        border-top-left-radius: 0;
        border-top-right-radius: 0;
    }
</style>
