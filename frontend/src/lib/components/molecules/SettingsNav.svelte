<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";

export let activeSection = "";

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
	}
}
</script>

<nav class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-4 sticky top-4 z-10">
	<h2 class="text-lg font-semibold mb-4 text-gray-900 dark:text-white">
		{$_("settings.nav.title")}
	</h2>

	<ul class="space-y-2">
		{#each navItems as item}
			<li>
				<div class="font-medium text-gray-800 dark:text-gray-200 py-1">
					{item.title}
				</div>
				{#if item.children}
					<ul class="ml-4 space-y-1">
						{#each item.children as child}
							<li>
								<button
									type="button"
									on:click={() => scrollToSection(child.id)}
									class="w-full text-left px-2 py-1 text-sm text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors cursor-pointer relative z-20 {activeSection === child.id ? 'text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20' : ''}"
									style="min-height: 32px; padding-top: 6px; padding-bottom: 6px;"
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