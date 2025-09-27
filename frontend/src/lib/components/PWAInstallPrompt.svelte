<script lang="ts">
import { onMount } from "svelte";
import { _ } from "svelte-i18n";
import "$lib/i18n";
import type { BeforeInstallPromptEvent } from "../pwa-types";

let deferredPrompt: BeforeInstallPromptEvent | null = null;
let showInstallPrompt = false;
let isInstallable = false;

onMount(() => {
	// Check if already installed
	if (window.matchMedia("(display-mode: standalone)").matches) {
		return;
	}

	// Check visit count
	const visitCount = Number(localStorage.getItem("pwa-visit-count") || "0") + 1;
	localStorage.setItem("pwa-visit-count", visitCount.toString());

	// Check if user has dismissed the prompt recently
	const dismissedAt = localStorage.getItem("pwa-prompt-dismissed");
	if (dismissedAt) {
		const dismissedTime = new Date(dismissedAt);
		const threeMonthsAgo = new Date();
		threeMonthsAgo.setMonth(threeMonthsAgo.getMonth() - 3);

		if (dismissedTime > threeMonthsAgo) {
			return;
		}
	}

	// Show prompt after 3 visits
	if (visitCount >= 3) {
		window.addEventListener("beforeinstallprompt", (e) => {
			e.preventDefault();
			deferredPrompt = e;
			isInstallable = true;
			showInstallPrompt = true;
		});
	}
});

const handleInstall = async () => {
	if (!deferredPrompt) return;

	showInstallPrompt = false;
	deferredPrompt.prompt();

	const { outcome } = await deferredPrompt.userChoice;

	if (outcome === "accepted") {
		localStorage.setItem("pwa-installed", "true");
	} else {
		localStorage.setItem("pwa-prompt-dismissed", new Date().toISOString());
	}

	deferredPrompt = null;
	isInstallable = false;
};

const handleDismiss = () => {
	showInstallPrompt = false;
	localStorage.setItem("pwa-prompt-dismissed", new Date().toISOString());
};
</script>

{#if showInstallPrompt && isInstallable}
	<div class="fixed bottom-4 left-4 right-4 z-50 max-w-sm mx-auto">
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-4">
			<div class="flex items-start space-x-3">
				<!-- App Icon -->
				<div class="flex-shrink-0">
					<img
						src="/icons/icon-72x72.png"
						alt="umi.mikan"
						class="w-12 h-12 rounded-lg"
					/>
				</div>

				<!-- Content -->
				<div class="flex-1 min-w-0">
					<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1">
						{$_("pwa.install.title")}
					</h3>
					<p class="text-xs text-gray-600 dark:text-gray-400 mb-3">
						{$_("pwa.install.description")}
					</p>

					<!-- Buttons -->
					<div class="flex space-x-2">
						<button
							on:click={handleInstall}
							class="text-xs bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-3 rounded-md transition-colors"
						>
							{$_("pwa.install.install")}
						</button>
						<button
							on:click={handleDismiss}
							class="text-xs bg-gray-200 dark:bg-gray-600 hover:bg-gray-300 dark:hover:bg-gray-500 text-gray-700 dark:text-gray-300 font-medium py-2 px-3 rounded-md transition-colors"
						>
							{$_("pwa.install.dismiss")}
						</button>
					</div>
				</div>

				<!-- Close Button -->
				<button
					on:click={handleDismiss}
					aria-label="Close install prompt"
					class="flex-shrink-0 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</button>
			</div>
		</div>
	</div>
{/if}