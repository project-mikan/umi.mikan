<script lang="ts">
import "../app.css";
import { page } from "$app/stores";
import "$lib/i18n";
import Head from "$lib/components/atoms/Head.svelte";
import NavigationBar from "$lib/components/molecules/NavigationBar.svelte";
import { _ } from "svelte-i18n";
import type { LayoutData } from "./$types";

export let data: LayoutData;

$: isAuthenticated = data.isAuthenticated;
$: isAuthPage =
	$page.url.pathname === "/login" || $page.url.pathname === "/register";
</script>

<Head />

<div class="min-h-screen bg-gray-50">
	<NavigationBar {isAuthenticated} {isAuthPage} />

	<main class="{isAuthenticated && !isAuthPage ? 'container mx-auto py-8' : ''}">
		<slot />
	</main>
</div>