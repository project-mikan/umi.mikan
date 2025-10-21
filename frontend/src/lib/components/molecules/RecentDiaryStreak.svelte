<script lang="ts">
import { _ } from "svelte-i18n";
import { goto } from "$app/navigation";
import "$lib/i18n";

interface DayStatus {
	date: string; // YYYY-MM-DD
	hasEntry: boolean;
	dayOfWeek: string;
	dayOfMonth: number;
}

export let recentDays: DayStatus[] = [];

// 日記ページに遷移
function navigateToDiary(dateStr: string) {
	goto(`/${dateStr}`);
}

// 曜日の短縮表記
function getWeekdayShort(dayOfWeek: string): string {
	// i18n対応: date.dayOfWeek.[曜日名]のキーを使用
	const key = `date.dayOfWeek.${dayOfWeek}`;
	return $_(`date.dayOfWeek.${dayOfWeek}`, {
		default: dayOfWeek.substring(0, 1),
	});
}
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-3">
	<h3 class="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-2 text-center">
		{$_("recentStreak.title")}
	</h3>
	<div class="grid grid-cols-7 gap-1">
		{#each recentDays as day}
			<button
				type="button"
				on:click={() => navigateToDiary(day.date)}
				class="flex flex-col items-center justify-center p-1 rounded transition-colors hover:bg-gray-100 dark:hover:bg-gray-700"
				title="{day.dayOfMonth}{$_('recentStreak.day')} ({getWeekdayShort(day.dayOfWeek)}) {day.hasEntry ? $_('recentStreak.written') : $_('recentStreak.notWritten')}"
			>
				<span class="text-xs text-gray-900 dark:text-gray-100 font-medium">
					{getWeekdayShort(day.dayOfWeek)}
				</span>
				<span class="text-base">
					{#if day.hasEntry}
						🔥
					{:else}
						<span class="text-gray-300 dark:text-gray-600">○</span>
					{/if}
				</span>
			</button>
		{/each}
	</div>
</div>
