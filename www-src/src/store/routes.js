import config from "./config";

import Home from "../views/Home.svelte";
import Admin from "../views/Admin.svelte"

export default {
    Logo: config.Logo,
    Routemap: [
        {route: "/", text: "Home", component: Home},
        {route: "/admin", text: "Admin Login", component: Admin, extrapadding:true},
    ]
}