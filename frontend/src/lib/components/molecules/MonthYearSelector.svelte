<script lang="ts">
import { _, locale } from "svelte-i18n";
import "$lib/i18n";

export let isOpen = false;
export let currentYear: number;
export let currentMonth: number;
export let onSelect: (year: number, month: number) => void;
export let onCancel: () => void;

let selectedYear = currentYear;
let selectedMonth = currentMonth;

$: {
	selectedYear = currentYear;
	selectedMonth = currentMonth;
}

function handleConfirm() {
	onSelect(selectedYear, selectedMonth);
}

function handleCancel() {
	onCancel();
}

const currentDate = new Date();
const minYear = 2020;
const maxYear = currentDate.getFullYear() + 5;

$: years = Array.from({ length: maxYear - minYear + 1 }, (_, i) => minYear + i);

$: months = Array.from({ length: 12 }, (_, i) => {
	const date = new Date(2000, i, 1);
	return {
		value: i + 1,
		label: date.toLocaleDateString($locale || "en", { month: "long" }),
	};
});
</script>

{#if isOpen}
	<div class="fixed inset-0 z-50 overflow-y-auto">
		<div class="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
			<!-- Backdrop -->
			<div class="fixed inset-0 transition-opacity" aria-hidden="true">
				<div class="absolute inset-0 bg-gray-500 dark:bg-gray-700 opacity-75" on:click={handleCancel} on:keydown={(e) => e.key === 'Escape' && handleCancel()} role="button" tabindex="-1"></div>
			</div>

			<span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

			<!-- Modal content -->
			<div class="inline-block align-bottom bg-white dark:bg-gray-800 rounded-lg text-left overflow-hidden shadow-xl dark:shadow-gray-900/20 transform transition-all sm:my-8 sm:align-middle sm:max-w-md sm:w-full">
				<div class="bg-white dark:bg-gray-800 px-4 pt-5 pb-4 sm:p-6">
					<div class="text-center">
						<h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100 mb-6">
							{$_("monthSelector.title")}
						</h3>
						
						<div class="space-y-4">
							<!-- Year selector -->
							<div>
								<label for="year-select" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
									{$_("monthSelector.year")}
								</label>
								<select 
									id="year-select"
									bind:value={selectedYear}
									class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								>
									{#each years as year}
										<option value={year}>{year}</option>
									{/each}
								</select>
							</div>

							<!-- Month selector -->
							<div>
								<label for="month-select" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
									{$_("monthSelector.month")}
								</label>
								<select 
									id="month-select"
									bind:value={selectedMonth}
									class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
								>
									{#each months as month}
										<option value={month.value}>{month.label}</option>
									{/each}
								</select>
							</div>
						</div>
					</div>
				</div>
				
				<!-- Action buttons -->
				<div class="bg-gray-50 dark:bg-gray-700 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
					<button
						on:click={handleConfirm}
						class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm"
					>
						{$_("monthSelector.confirm")}
					</button>
					<button
						on:click={handleCancel}
						class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 dark:border-gray-600 shadow-sm px-4 py-2 bg-white dark:bg-gray-800 text-base font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
					>
						{$_("monthSelector.cancel")}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}