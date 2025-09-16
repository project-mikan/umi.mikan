<script lang="ts">
import { _ } from "svelte-i18n";
import type { DiaryEntry } from "$lib/grpc";
import type { DateInfo } from "$lib/utils/date-utils";
import { formatDateToId } from "$lib/utils/date-utils";
import Link from "../atoms/Link.svelte";

export let pastEntries: {
	oneWeekAgo: { date: DateInfo; entry: DiaryEntry | null };
	oneMonthAgo: { date: DateInfo; entry: DiaryEntry | null };
	twoMonthsAgo: { date: DateInfo; entry: DiaryEntry | null };
	sixMonthsAgo: { date: DateInfo; entry: DiaryEntry | null };
	oneYearAgo: { date: DateInfo; entry: DiaryEntry | null };
	twoYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
	threeYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
	fourYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
	fiveYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
	sixYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
	sevenYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
	eightYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
	nineYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
	tenYearsAgo: { date: DateInfo; entry: DiaryEntry | null };
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
		date: pastEntries.twoMonthsAgo.date,
		labelKey: "diary.twoMonthsAgo",
		entry: pastEntries.twoMonthsAgo.entry,
	},
	{
		date: pastEntries.sixMonthsAgo.date,
		labelKey: "diary.sixMonthsAgo",
		entry: pastEntries.sixMonthsAgo.entry,
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
	{
		date: pastEntries.threeYearsAgo.date,
		labelKey: "diary.threeYearsAgo",
		entry: pastEntries.threeYearsAgo.entry,
	},
	{
		date: pastEntries.fourYearsAgo.date,
		labelKey: "diary.fourYearsAgo",
		entry: pastEntries.fourYearsAgo.entry,
	},
	{
		date: pastEntries.fiveYearsAgo.date,
		labelKey: "diary.fiveYearsAgo",
		entry: pastEntries.fiveYearsAgo.entry,
	},
	{
		date: pastEntries.sixYearsAgo.date,
		labelKey: "diary.sixYearsAgo",
		entry: pastEntries.sixYearsAgo.entry,
	},
	{
		date: pastEntries.sevenYearsAgo.date,
		labelKey: "diary.sevenYearsAgo",
		entry: pastEntries.sevenYearsAgo.entry,
	},
	{
		date: pastEntries.eightYearsAgo.date,
		labelKey: "diary.eightYearsAgo",
		entry: pastEntries.eightYearsAgo.entry,
	},
	{
		date: pastEntries.nineYearsAgo.date,
		labelKey: "diary.nineYearsAgo",
		entry: pastEntries.nineYearsAgo.entry,
	},
	{
		date: pastEntries.tenYearsAgo.date,
		labelKey: "diary.tenYearsAgo",
		entry: pastEntries.tenYearsAgo.entry,
	},
];

function _getEntryTitle(entry: DiaryEntry | null): string {
	if (!entry || !entry.content) return $_("diary.noPastEntry");

	// 最初の30文字を取得してタイトルとする
	const firstLine = entry.content.split("\n")[0];
	return firstLine.length > 30 ? `${firstLine.substring(0, 30)}...` : firstLine;
}
</script>

<div class="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700">
	<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">
		{$_('diary.pastEntries')}
	</h3>
	
	<div class="space-y-3">
		{#each pastEntriesList as pastEntry}
			<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
				<div class="flex-1">
					<div class="text-sm text-gray-600 dark:text-gray-400 mb-1">
						{$_(pastEntry.labelKey)} ({$_('date.format.yearMonthDay', {
							values: {
								year: pastEntry.date.year,
								month: pastEntry.date.month,
								day: pastEntry.date.day
							}
						})})
					</div>
					<div class="text-sm text-gray-800 dark:text-gray-200">
						{#if pastEntry.entry}
							<Link href="/{formatDateToId(pastEntry.date)}" class="text-blue-600 hover:text-blue-800">
								{_getEntryTitle(pastEntry.entry)}
							</Link>
						{:else}
							<span class="text-gray-400 dark:text-gray-500">{$_('diary.noPastEntry')}</span>
						{/if}
					</div>
				</div>
			</div>
		{/each}
	</div>
</div>