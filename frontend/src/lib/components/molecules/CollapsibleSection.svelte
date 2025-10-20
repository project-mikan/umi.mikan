<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";

export let title: string;
export let isOpen = false;

function toggleSection() {
	isOpen = !isOpen;
}
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 mb-4">
	<button
		type="button"
		on:click={toggleSection}
		class="w-full flex items-center justify-between px-4 py-3 text-left hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors rounded-lg"
	>
		<h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">
			{title}
		</h3>
		<svg
			class="w-5 h-5 text-gray-500 dark:text-gray-400 transition-transform {isOpen
				? 'rotate-180'
				: ''}"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M19 9l-7 7-7-7"
			/>
		</svg>
	</button>

	{#if isOpen}
		<div class="px-4 pb-4">
			<slot />
		</div>
	{/if}
</div>
