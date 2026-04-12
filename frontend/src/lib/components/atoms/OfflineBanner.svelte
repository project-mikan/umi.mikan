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
		class="fixed bottom-0 left-0 right-0 z-50 bg-yellow-500 dark:bg-yellow-600 text-white px-4 py-2 flex items-center justify-center gap-2 text-sm shadow-lg"
	>
		<svg
			class="w-4 h-4 flex-shrink-0"
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
		<span>{$_("offline.banner.message")}</span>
	</div>
{/if}
