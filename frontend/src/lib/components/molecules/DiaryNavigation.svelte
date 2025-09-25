<script lang="ts">
import { _ } from "svelte-i18n";
import { goto, invalidateAll } from "$app/navigation";
import type { DateInfo } from "$lib/utils/date-utils";
import {
	formatDateToId,
	getDayOfWeekKey,
	getNextDate,
	getPreviousDate,
} from "$lib/utils/date-utils";
import Button from "../atoms/Button.svelte";

export let currentDate: DateInfo;

async function _goToPreviousDay() {
	const previousDate = getPreviousDate(currentDate);
	const id = formatDateToId(previousDate);
	await goto(`/${id}`);
	await invalidateAll();
}

async function _goToNextDay() {
	const nextDate = getNextDate(currentDate);
	const id = formatDateToId(nextDate);
	await goto(`/${id}`);
	await invalidateAll();
}
</script>

<div class="flex items-center justify-between">
	<Button
		variant="secondary"
		size="sm"
		on:click={_goToPreviousDay}
		class="flex items-center gap-2"
	>
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M15 19l-7-7 7-7"
			/>
		</svg>
		{$_("diary.previousDay")}
	</Button>

	<div class="text-center">
		<span class="text-xl font-semibold text-gray-700 dark:text-gray-300">
			{$_("date.format.yearMonthDayWithDayOfWeek", {
				values: {
					year: currentDate.year,
					month: currentDate.month,
					day: currentDate.day,
					dayOfWeek: $_(`date.dayOfWeek.${getDayOfWeekKey(currentDate)}`),
				},
			})}
		</span>
	</div>

	<Button
		variant="secondary"
		size="sm"
		on:click={_goToNextDay}
		class="flex items-center gap-2"
	>
		{$_("diary.nextDay")}
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M9 5l7 7-7 7"
			/>
		</svg>
	</Button>
</div>
