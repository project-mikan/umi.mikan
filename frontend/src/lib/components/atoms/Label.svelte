<script lang="ts">
export let htmlFor = "";
export let required = false;
export let srOnly = false;

$: classes = `${srOnly ? "sr-only" : "block text-sm font-medium text-gray-700 dark:text-gray-300"}`;
</script>

<label for={htmlFor} class={classes}>
	<slot />
	{#if required && !srOnly}
		<span class="text-red-500 dark:text-red-400">*</span>
	{/if}
</label>