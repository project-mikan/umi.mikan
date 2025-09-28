<script lang="ts">
import { onMount } from "svelte";
import { _ } from "svelte-i18n";
import "$lib/i18n";

let showInstallButton = false;
let installPrompt: BeforeInstallPromptEvent | null = null;

interface BeforeInstallPromptEvent extends Event {
	prompt(): Promise<void>;
	userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

onMount(() => {
	// Listen for beforeinstallprompt event
	window.addEventListener("beforeinstallprompt", (e) => {
		e.preventDefault();
		installPrompt = e as BeforeInstallPromptEvent;
		showInstallButton = true;
	});

	// Hide button if app is already installed
	window.addEventListener("appinstalled", () => {
		showInstallButton = false;
		installPrompt = null;
	});

	// For development/testing: show button if no install prompt is available
	// In production, this will be overridden by the beforeinstallprompt event
	setTimeout(() => {
		if (!installPrompt) {
			showInstallButton = true;
		}
	}, 1000);
});

async function handleInstall() {
	if (!installPrompt) {
		// Show manual installation instructions if no automatic prompt is available
		alert($_("pwa.install.manualInstructions"));
		return;
	}

	try {
		await installPrompt.prompt();
		const choiceResult = await installPrompt.userChoice;

		if (choiceResult.outcome === "accepted") {
			showInstallButton = false;
		}

		installPrompt = null;
	} catch (error) {
		console.error("Install failed:", error);
	}
}
</script>

{#if showInstallButton}
	<div class="mt-8 text-center">
		<button
			on:click={handleInstall}
			class="inline-flex items-center px-4 py-2 text-sm font-medium text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 border border-blue-600 dark:border-blue-400 rounded-md hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
		>
			<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8l-8-8-8 8"/>
			</svg>
			{$_("pwa.install.install")}
		</button>
	</div>
{/if}