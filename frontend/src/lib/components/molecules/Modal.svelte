<script lang="ts">
import Button from "../atoms/Button.svelte";

export let isOpen = false;
export let title: string;
export let confirmText: string;
export let cancelText: string;
export let variant: "danger" | "primary" = "primary";
export let onConfirm: (() => void) | null = null;
export let onCancel: (() => void) | null = null;

const iconPaths = {
	danger:
		"M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z",
	primary: "M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z",
};

const iconColors = {
	danger: "text-red-600",
	primary: "text-blue-600",
};

const iconBackgrounds = {
	danger: "bg-red-100",
	primary: "bg-blue-100",
};

function handleConfirm() {
	if (onConfirm) {
		onConfirm();
	}
}

function handleCancel() {
	if (onCancel) {
		onCancel();
	}
}
</script>

{#if isOpen}
	<div class="fixed inset-0 z-50 overflow-y-auto">
		<div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
			<div class="fixed inset-0 transition-opacity" aria-hidden="true">
				<div class="absolute inset-0 bg-gray-500 opacity-75"></div>
			</div>

			<span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

			<div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
				<div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
					<div class="sm:flex sm:items-start">
						<div class="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full {iconBackgrounds[variant]} sm:mx-0 sm:h-10 sm:w-10">
							<svg class="h-6 w-6 {iconColors[variant]}" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d={iconPaths[variant]} />
							</svg>
						</div>
						<div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
							<h3 class="text-lg leading-6 font-medium text-gray-900">
								{title}
							</h3>
							<div class="mt-2">
								<slot />
							</div>
						</div>
					</div>
				</div>
				<div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
					<Button
						variant={variant}
						size="sm"
						on:click={handleConfirm}
						class="w-full sm:ml-3 sm:w-auto"
					>
						{confirmText}
					</Button>
					<Button
						variant="secondary"
						size="sm"
						on:click={handleCancel}
						class="mt-3 w-full sm:mt-0 sm:ml-3 sm:w-auto"
					>
						{cancelText}
					</Button>
				</div>
			</div>
		</div>
	</div>
{/if}