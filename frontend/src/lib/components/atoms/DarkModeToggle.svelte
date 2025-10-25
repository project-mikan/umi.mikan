<script lang="ts">
	import { browser } from "$app/environment";
	import { onMount } from "svelte";

	let isDarkMode = false;

	// Initialize dark mode from localStorage or system preference
	onMount(() => {
		if (browser) {
			const stored = localStorage.getItem("darkMode");
			if (stored !== null) {
				isDarkMode = stored === "true";
			} else {
				// Use system preference
				const systemDark = window.matchMedia(
					"(prefers-color-scheme: dark)",
				).matches;
				isDarkMode = systemDark;
			}
			applyDarkMode();

			// Listen to system color scheme changes
			const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
			mediaQuery.addEventListener("change", (e) => {
				// Only apply system preference if user hasn't set a manual preference
				if (localStorage.getItem("darkMode") === null) {
					isDarkMode = e.matches;
					applyDarkMode();
				}
			});
		}
	});

	function toggleDarkMode() {
		isDarkMode = !isDarkMode;
		if (browser) {
			localStorage.setItem("darkMode", isDarkMode.toString());
		}
		applyDarkMode();
	}

	function applyDarkMode() {
		if (browser) {
			if (isDarkMode) {
				document.documentElement.classList.add("dark");
			} else {
				document.documentElement.classList.remove("dark");
			}
		}
	}
</script>

<button
	type="button"
	class="flex items-center px-3 py-2 text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800 rounded-md transition-colors"
	on:click={toggleDarkMode}
	aria-label={isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'}
	title={isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'}
>
	{#if isDarkMode}
		<!-- Sun icon for light mode -->
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"></path>
		</svg>
	{:else}
		<!-- Moon icon for dark mode -->
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path>
		</svg>
	{/if}
	<span class="sr-only">{isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'}</span>
</button>