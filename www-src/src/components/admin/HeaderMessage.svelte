<script>
    import { fade } from "svelte/transition";
    export let success = true
    let border = "border-success"
    if (!success) {
        border = "border-danger"
    }
</script>

<div class="border border-3 mt-3 rounded-3 p-3 {border}" transition:fade>
    <slot/>
</div>