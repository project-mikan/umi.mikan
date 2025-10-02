<script lang="ts">
	import { _ } from "svelte-i18n";
	import { browser } from "$app/environment";
	import "$lib/i18n";
	import { budouxEnabled } from "$lib/budoux-store";

	function toggleBudouX() {
		budouxEnabled.toggle();
		// ページをリロードして設定を反映
		if (browser) {
			window.location.reload();
		}
	}
</script>

<button
	type="button"
	on:click={toggleBudouX}
	class="p-2 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors relative group"
	aria-label={$_("budoux.toggle")}
	title={$_("budoux.description")}
>
	{#if $budouxEnabled}
		<!-- BudouX有効時のアイコン -->
		<svg
			class="w-6 h-6 text-blue-600 dark:text-blue-400"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
			/>
		</svg>
	{:else}
		<!-- BudouX無効時のアイコン -->
		<svg
			class="w-6 h-6 text-gray-600 dark:text-gray-400"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
			/>
		</svg>
	{/if}
</button>
