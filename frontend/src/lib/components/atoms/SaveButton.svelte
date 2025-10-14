<script lang="ts">
import { _ } from "svelte-i18n";
import Button from "./Button.svelte";

export let loading = false;
export let saved = false;
export let size: "sm" | "md" | "lg" = "md";
</script>

<Button type="submit" variant={saved ? "success" : "primary"} {size} disabled={loading || saved} on:click>
	<div class="flex items-center justify-center min-h-[1.25rem]">
		{#if loading}
			<svg class="animate-spin -mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
				<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 714 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
			</svg>
			<span class="ml-1">{$_("diary.saving")}</span>
		{:else if saved}
			<svg class="-mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
				<path stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" d="m9 12 2 2 4-4"/>
			</svg>
			<span class="ml-1">{$_("diary.saved")}</span>
		{:else}
			<span>{$_("diary.save")}</span>
		{/if}
	</div>
</Button>