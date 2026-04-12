<script lang="ts">
  import { _ } from "svelte-i18n";
  import "$lib/i18n";
  import { page } from "$app/stores";
  import { isOnline } from "$lib/online-store";
</script>

{#if !$isOnline && $page.url.pathname !== "/offline"}
	<div
		role="status"
		aria-live="polite"
		class="fixed top-16 right-4 z-50 bg-gray-800 dark:bg-gray-700 text-white px-3 py-2 rounded-lg shadow-lg flex items-center gap-2 text-xs"
	>
		<svg
			class="w-3.5 h-3.5 flex-shrink-0 text-yellow-400"
			fill="none"
			stroke="currentColor"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M18.364 5.636l-12.728 12.728m0 0L5.636 18.364m12.728-12.728L18.364 18.364M12 12h.01"
			/>
		</svg>
		<span>{$_("offline.title")}</span>
	</div>
{/if}
