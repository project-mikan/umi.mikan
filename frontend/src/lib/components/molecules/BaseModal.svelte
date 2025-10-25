<script lang="ts">
	import { onMount } from "svelte";

	export let isOpen = false;
	export let onBackdropClick: (() => void) | null = null;
	export let maxWidth = "max-w-lg";

	function _handleBackdropClick() {
		if (onBackdropClick) {
			onBackdropClick();
		}
	}

	function _handleKeydown(event: KeyboardEvent) {
		if (event.key === "Escape" && onBackdropClick) {
			onBackdropClick();
		}
	}

	onMount(() => {
		const handleGlobalKeydown = (event: KeyboardEvent) => {
			if (isOpen && event.key === "Escape" && onBackdropClick) {
				onBackdropClick();
			}
		};

		document.addEventListener("keydown", handleGlobalKeydown);

		return () => {
			document.removeEventListener("keydown", handleGlobalKeydown);
		};
	});
</script>

{#if isOpen}
	<div class="fixed inset-0 z-[9999] overflow-y-auto">
		<div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
			<!-- Backdrop -->
			<button 
				class="fixed inset-0 z-[9998] transition-opacity bg-gray-500 dark:bg-gray-700 opacity-75" 
				on:click={_handleBackdropClick} 
				on:keydown={_handleKeydown}
				aria-label="Close modal"
			></button>

			<span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

			<!-- Modal content -->
			<div class="inline-block align-bottom bg-white dark:bg-gray-800 rounded-lg text-left overflow-hidden shadow-xl dark:shadow-gray-900/20 transform transition-all sm:my-8 sm:align-middle {maxWidth} sm:w-full z-[10000] relative">
				<slot />
			</div>
		</div>
	</div>
{/if}