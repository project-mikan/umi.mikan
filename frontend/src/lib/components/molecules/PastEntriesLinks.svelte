<script lang="ts">
import Link from "$lib/components/atoms/Link.svelte";
import type { DiaryEntry } from "$lib/grpc";
import { formatDateToId } from "$lib/utils/date-utils";
import type { DateInfo } from "$lib/utils/date-utils";
import { _ } from "svelte-i18n";

export let pastEntries: {
	oneWeekAgo: { date: DateInfo; entry: DiaryEntry | null };
	oneMonthAgo: { date: DateInfo; entry: DiaryEntry | null };
	oneYearAgo: { date: DateInfo; entry: DiaryEntry | null };
	twoYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
};

interface PastEntry {
	date: DateInfo;
	labelKey: string;
	entry: DiaryEntry | null;
}

$: pastEntriesList = [
	{
		date: pastEntries.oneWeekAgo.date,
		labelKey: "diary.oneWeekAgo",
		entry: pastEntries.oneWeekAgo.entry,
	},
	{
		date: pastEntries.oneMonthAgo.date,
		labelKey: "diary.oneMonthAgo",
		entry: pastEntries.oneMonthAgo.entry,
	},
	{
		date: pastEntries.oneYearAgo.date,
		labelKey: "diary.oneYearAgo",
		entry: pastEntries.oneYearAgo.entry,
	},
	{
		date: pastEntries.twoYearsAgo.date,
		labelKey: "diary.twoYearsAgo",
		entry: pastEntries.twoYearsAgo.entry,
	},
];

function getEntryTitle(entry: DiaryEntry | null): string {
	if (!entry || !entry.content) return $_("diary.noPastEntry");

	// 最初の30文字を取得してタイトルとする
	const firstLine = entry.content.split("\n")[0];
	return firstLine.length > 30 ? `${firstLine.substring(0, 30)}...` : firstLine;
}
</script>

<div class="mt-8 pt-6 border-t border-gray-200">
	<h3 class="text-lg font-semibold text-gray-900 mb-4">
		{$_('diary.pastEntries')}
	</h3>
	
	<div class="space-y-3">
		{#each pastEntriesList as pastEntry}
			<div class="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
				<div class="flex-1">
					<div class="text-sm text-gray-600 mb-1">
						{$_(pastEntry.labelKey)} ({$_('date.format.yearMonthDay', {
							values: {
								year: pastEntry.date.year,
								month: pastEntry.date.month,
								day: pastEntry.date.day
							}
						})})
					</div>
					<div class="text-sm text-gray-800">
						{#if pastEntry.entry}
							<Link href="/{formatDateToId(pastEntry.date)}" class="text-blue-600 hover:text-blue-800">
								{getEntryTitle(pastEntry.entry)}
							</Link>
						{:else}
							<span class="text-gray-400">{$_('diary.noPastEntry')}</span>
						{/if}
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>