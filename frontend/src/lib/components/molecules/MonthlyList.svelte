<script lang="ts">
import { _ } from "svelte-i18n";
import type { DiaryEntry } from "$lib/grpc";

export let daysInMonth: number;
export let currentYear: number;
export let currentMonth: number;
export let entryMap: Map<number, DiaryEntry>;
export let onNavigateToEntry: (day: number) => void;

function formatContentWithLineBreaks(content: string): string {
	return content.replace(/\n/g, "<br>");
}
</script>

<div class="space-y-3">
	{#each Array.from({ length: daysInMonth }, (_, i) => i + 1) as day}
		{@const entry = entryMap.get(day)}
		<div class="bg-white dark:bg-gray-800 shadow dark:shadow-gray-900/20 rounded-lg overflow-hidden">
			<button
				on:click={() => onNavigateToEntry(day)}
				class="w-full p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer"
			>
				<div class="flex items-center justify-between mb-2">
					<div class="flex items-center space-x-3">
						<span class="text-lg font-semibold text-gray-900 dark:text-gray-100">{day}</span>
						<span class="text-sm text-gray-500 dark:text-gray-400">
							{new Date(currentYear, currentMonth - 1, day).toLocaleDateString(undefined, { weekday: 'short' })}
						</span>
					</div>
					{#if !entry}
						<span class="text-blue-600 dark:text-blue-400 text-sm font-medium">+ {$_("monthly.entry")}</span>
					{/if}
				</div>
				
				{#if entry}
					<div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-3">
						<div class="text-sm text-blue-800 dark:text-blue-300 font-medium mb-2">
							{$_("monthly.entry")}
						</div>
						<div class="text-sm text-blue-600 dark:text-blue-400 leading-relaxed">
							{@html formatContentWithLineBreaks(
								entry?.content
									? entry.content.substring(0, 100) +
											(entry.content.length > 100 ? "..." : "")
									: "",
							)}
						</div>
					</div>
				{:else}
					<div class="text-gray-400 dark:text-gray-500 text-sm italic">
						{$_("monthly.noEntry", { default: "日記がありません" })}
					</div>
				{/if}
			</button>
		</div>
	{/each}
</div>