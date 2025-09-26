<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";

export let activeSection = "";
export let isOpen = false;
export let onToggle: () => void;

interface NavItem {
	id: string;
	title: string;
	children?: NavItem[];
}

let isScrolling = false;
let scrollTimeout: number;

$: navItems = [
	{
		id: "user-settings",
		title: $_("settings.nav.userSettings"),
		children: [
			{ id: "username", title: $_("settings.username.title") },
			{ id: "password", title: $_("settings.password.title") },
		],
	},
	{
		id: "llm-settings",
		title: $_("settings.nav.llmSettings"),
		children: [
			{ id: "llm-token", title: $_("settings.llmToken.title") },
			{ id: "auto-summary", title: $_("settings.autoSummary.title") },
			{ id: "llm-status", title: $_("settings.llmStatus.title") },
		],
	},
	{
		id: "danger-zone",
		title: $_("settings.nav.dangerZone"),
		children: [
			{ id: "delete-account", title: $_("settings.deleteAccount.title") },
		],
	},
];

function scrollToSection(sectionId: string) {
	const element = document.getElementById(sectionId);
	if (element) {
		// Mark as scrolling to prevent intersection observer conflicts
		isScrolling = true;
		activeSection = sectionId; // Set immediately for better UX

		// Clear any existing timeout
		if (scrollTimeout) {
			clearTimeout(scrollTimeout);
		}

		element.scrollIntoView({ behavior: "smooth", block: "start" });

		// Reset scrolling flag after scroll completes
		scrollTimeout = setTimeout(() => {
			isScrolling = false;
		}, 1000);

		// Close mobile menu after navigation
		if (onToggle && window.innerWidth < 768) {
			onToggle();
		}
	}
}
</script>

<!-- Mobile Navigation Overlay -->
<div
	class="md:hidden fixed inset-0 bg-black bg-opacity-50 z-40 transition-opacity duration-300 {isOpen ? 'opacity-100' : 'opacity-0 pointer-events-none'}"
	on:click={onToggle}
></div>

<!-- Navigation Sidebar -->
<nav class="bg-white dark:bg-gray-800 shadow-md z-50
	md:block md:relative md:transform-none md:opacity-100 md:w-auto md:rounded-lg md:p-4 md:sticky md:top-4
	{isOpen ? 'fixed top-0 left-0 h-full w-full transform translate-x-0 p-6 overflow-y-auto' : 'hidden md:block md:translate-x-0'}
	transition-transform duration-300 ease-in-out md:transition-none">

	<!-- Mobile Header with Close Button -->
	<div class="md:hidden flex justify-between items-center mb-6 pb-4 border-b border-gray-200 dark:border-gray-700">
		<h2 class="text-xl font-semibold text-gray-900 dark:text-white">
			{$_("settings.nav.title")}
		</h2>
		<button
			type="button"
			on:click={onToggle}
			class="p-2 rounded-md text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
		>
			<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
			</svg>
		</button>
	</div>

	<!-- Desktop Header -->
	<h2 class="hidden md:block text-lg font-semibold mb-4 text-gray-900 dark:text-white">
		{$_("settings.nav.title")}
	</h2>

	<ul class="space-y-3 md:space-y-2">
		{#each navItems as item}
			<li>
				<div class="font-medium text-gray-800 dark:text-gray-200 py-2 text-base md:text-sm md:py-1">
					{item.title}
				</div>
				{#if item.children}
					<ul class="ml-4 space-y-2 md:space-y-1">
						{#each item.children as child}
							<li>
								<button
									type="button"
									on:click={() => scrollToSection(child.id)}
									class="w-full text-left px-3 py-3 md:px-2 md:py-1 text-base md:text-sm text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors cursor-pointer relative z-20 {activeSection === child.id ? 'text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20' : ''}"
									style="min-height: 44px; md:min-height: 32px;"
								>
									{child.title}
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</li>
		{/each}
	</ul>
</nav>