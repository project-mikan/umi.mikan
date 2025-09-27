<script lang="ts">
import "../app.css";
import "$lib/i18n";
import { page } from "$app/stores";
import { onMount } from "svelte";
import Head from "$lib/components/atoms/Head.svelte";
import NavigationBar from "$lib/components/molecules/NavigationBar.svelte";
import QuickNavigation from "$lib/components/molecules/QuickNavigation.svelte";
import Footer from "$lib/components/organisms/Footer.svelte";
import { summaryVisibility } from "$lib/summary-visibility-store";
import PWAInstallPrompt from "$lib/components/PWAInstallPrompt.svelte";
import PWAUpdateNotification from "$lib/components/PWAUpdateNotification.svelte";
import type { LayoutData } from "./$types";

export let data: LayoutData;

$: isAuthenticated = data.isAuthenticated;
$: isAuthPage =
	$page.url.pathname === "/login" || $page.url.pathname === "/register";

onMount(() => {
	// ストアを初期化（アプリケーション全体で一度だけ）
	summaryVisibility.init();
});
</script>

<Head />

<div class="min-h-screen bg-gray-50 dark:bg-gray-800 transition-colors flex flex-col">
	<NavigationBar {isAuthenticated} {isAuthPage} />

	{#if isAuthenticated && !isAuthPage}
		<QuickNavigation />
	{/if}

	<main class="{isAuthenticated && !isAuthPage ? 'container mx-auto py-8' : ''} text-gray-900 dark:text-gray-100 flex-1">
		<slot />
	</main>

	<Footer {isAuthenticated} />
</div>

<!-- PWA Components -->
<PWAInstallPrompt />
<PWAUpdateNotification />