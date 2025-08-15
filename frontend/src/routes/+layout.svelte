<script lang="ts">
import "../app.css";
import "$lib/i18n";
import { page } from "$app/stores";
import { onMount } from "svelte";
import Head from "$lib/components/atoms/Head.svelte";
import NavigationBar from "$lib/components/molecules/NavigationBar.svelte";
import QuickNavigation from "$lib/components/molecules/QuickNavigation.svelte";
import type { LayoutData } from "./$types";

export let data: LayoutData;

$: isAuthenticated = data.isAuthenticated;
$: isAuthPage =
	$page.url.pathname === "/login" || $page.url.pathname === "/register";

onMount(() => {
	// Layout initialization
});
</script>

<Head />

<div class="min-h-screen bg-gray-50 dark:bg-gray-800 transition-colors">
	<NavigationBar {isAuthenticated} {isAuthPage} />

	{#if isAuthenticated && !isAuthPage}
		<QuickNavigation />
	{/if}

	<main class="{isAuthenticated && !isAuthPage ? 'container mx-auto py-8' : ''} text-gray-900 dark:text-gray-100">
		<slot />
	</main>
</div>