<script lang="ts">
import { _ } from "svelte-i18n";
import type { DiaryEntry } from "$lib/grpc/diary/diary_pb.ts";

export let calendarDays: (number | null)[];
export let entryMap: Map<number, DiaryEntry>;
export let weekDays: string[];
export let onNavigateToEntry: (day: number) => void;

function formatContentWithLineBreaks(content: string): string {
	return content.replace(/\n/g, "<br>");
}
</script>

<div class="bg-white dark:bg-gray-800 shadow dark:shadow-gray-900/20 rounded-lg overflow-hidden">
	<!-- 曜日ヘッダー -->
	<div class="grid grid-cols-7 bg-gray-50 dark:bg-gray-700">
		{#each weekDays as weekDay}
			<div
				class="p-4 text-center font-medium text-gray-700 dark:text-gray-300 border-r border-gray-200 dark:border-gray-600 last:border-r-0"
			>
				{weekDay}
			</div>
		{/each}
	</div>

	<!-- カレンダーグリッド -->
	<div class="grid grid-cols-7">
		{#each calendarDays as day}
			<div class="h-32 border-r border-b border-gray-200 dark:border-gray-600 last:border-r-0">
				{#if day !== null}
					<button
						on:click={() => onNavigateToEntry(day)}
						class="w-full h-full p-2 text-left hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors cursor-pointer"
					>
						<div class="h-full flex flex-col">
							<!-- 日付 -->
							<div class="flex justify-between items-start mb-1">
								<span class="text-sm font-medium text-gray-700 dark:text-gray-300">{day}</span>
								{#if !entryMap.has(day)}
									<span class="text-xs text-blue-600 dark:text-blue-400 opacity-50"> + </span>
								{/if}
							</div>

							<!-- 日記エントリ -->
							{#if entryMap.has(day)}
								{@const entry = entryMap.get(day)}
								<div class="flex-1 min-h-0">
									<div class="bg-blue-50 dark:bg-blue-900/20 rounded p-2 h-full">
										<div class="text-xs text-blue-800 dark:text-blue-300 font-medium mb-1">
											{$_("monthly.entry")}
										</div>
										<div class="text-xs text-blue-600 dark:text-blue-400 line-clamp-2">
											{@html formatContentWithLineBreaks(
												entry?.content
													? entry.content.substring(0, 40) +
															(entry.content.length > 40 ? "..." : "")
													: "",
											)}
										</div>
									</div>
								</div>
							{/if}
						</div>
					</button>
				{/if}
			</div>
		{/each}
	</div>
</div>

<style>
	.line-clamp-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>