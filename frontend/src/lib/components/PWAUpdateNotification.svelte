<script lang="ts">
import { onMount } from "svelte";
import { _ } from "svelte-i18n";
import "$lib/i18n";
let pwaInfo: unknown = null;

let showUpdatePrompt = false;
let updateServiceWorker: (() => Promise<void>) | null = null;

onMount(async () => {
	try {
		// @ts-expect-error
		const pwaModule = await import("virtual:pwa-info");
		pwaInfo = pwaModule.pwaInfo;

		if (pwaInfo) {
			// @ts-expect-error
			const { registerSW } = await import("virtual:pwa-register");

			updateServiceWorker = registerSW({
				immediate: true,
				onNeedRefresh() {
					showUpdatePrompt = true;
				},
				onOfflineReady() {
					console.log("App is ready to work offline");
				},
			});
		}
	} catch (error) {
		console.warn("PWA modules not available:", error);
	}
});

const handleUpdate = async () => {
	if (updateServiceWorker) {
		try {
			await updateServiceWorker();
			showUpdatePrompt = false;
			window.location.reload();
		} catch (error) {
			console.error("Failed to update service worker:", error);
		}
	}
};

const handleDismiss = () => {
	showUpdatePrompt = false;
};
</script>

{#if showUpdatePrompt}
	<div class="fixed top-4 left-4 right-4 z-50 max-w-md mx-auto">
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 p-4">
			<div class="flex items-start space-x-3">
				<!-- Update Icon -->
				<div class="flex-shrink-0">
					<svg
						class="w-6 h-6 text-blue-500"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
						/>
					</svg>
				</div>

				<!-- Content -->
				<div class="flex-1 min-w-0">
					<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1">
						{$_("pwa.update.title")}
					</h3>
					<p class="text-xs text-gray-600 dark:text-gray-400 mb-3">
						{$_("pwa.update.description")}
					</p>

					<!-- Buttons -->
					<div class="flex space-x-2">
						<button
							on:click={handleUpdate}
							class="text-xs bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-3 rounded-md transition-colors"
						>
							{$_("pwa.update.update")}
						</button>
						<button
							on:click={handleDismiss}
							class="text-xs bg-gray-200 dark:bg-gray-600 hover:bg-gray-300 dark:hover:bg-gray-500 text-gray-700 dark:text-gray-300 font-medium py-2 px-3 rounded-md transition-colors"
						>
							{$_("pwa.update.later")}
						</button>
					</div>
				</div>

				<!-- Close Button -->
				<button
					on:click={handleDismiss}
					aria-label="Close update notification"
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