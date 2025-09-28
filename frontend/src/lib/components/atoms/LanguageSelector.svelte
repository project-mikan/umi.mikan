<script lang="ts">
import { locale } from "svelte-i18n";
import { browser } from "$app/environment";

let isOpen = false;

const languages = [
	{ code: "en", label: "English" },
	{ code: "ja", label: "日本語" },
];

function toggleDropdown() {
	isOpen = !isOpen;
}

function selectLanguage(langCode: string) {
	if (browser) {
		locale.set(langCode);
		localStorage.setItem("locale", langCode);

		// Update manifest link for new language
		updateManifest(langCode);
	}
	isOpen = false;
}

function updateManifest(lang: string) {
	if (!browser) return;

	// Remove existing manifest link
	const existingManifest = document.querySelector('link[rel="manifest"]');
	if (existingManifest) {
		existingManifest.remove();
	}

	// Add new manifest link with updated language
	const manifestLink = document.createElement("link");
	manifestLink.rel = "manifest";
	manifestLink.href = `/manifest.webmanifest?lang=${lang}`;
	document.head.appendChild(manifestLink);
}

function closeDropdown() {
	isOpen = false;
}

// Close dropdown when clicking outside
function handleClickOutside(event: MouseEvent) {
	const target = event.target as Element;
	if (!target.closest(".language-selector")) {
		isOpen = false;
	}
}

$: if (browser && isOpen) {
	document.addEventListener("click", handleClickOutside);
} else if (browser) {
	document.removeEventListener("click", handleClickOutside);
}

$: currentLanguage =
	languages.find((lang) => lang.code === $locale) || languages[0];
</script>

<div class="language-selector relative">
	<button
		type="button"
		class="flex items-center px-3 py-2 text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800 rounded-md transition-colors"
		on:click={toggleDropdown}
		aria-expanded={isOpen}
		aria-haspopup="true"
	>
		<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129"></path>
		</svg>
		<span>{currentLanguage.label}</span>
		<svg class="w-4 h-4 ml-1 transition-transform {isOpen ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
		</svg>
	</button>

	{#if isOpen}
		<div class="absolute right-0 mt-2 w-40 bg-white dark:bg-gray-800 rounded-md shadow-lg border border-gray-200 dark:border-gray-600 z-50">
			<div class="py-1">
				{#each languages as language}
					<button
						type="button"
						class="w-full text-left px-4 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 focus:outline-none focus:bg-gray-100 dark:focus:bg-gray-700 {language.code === $locale ? 'bg-blue-50 dark:bg-blue-900 text-blue-700 dark:text-blue-300' : ''}"
						on:click={() => selectLanguage(language.code)}
					>
						{language.label}
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>