<script lang="ts">
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import { _ } from "svelte-i18n";
import "$lib/i18n";
import type { DiaryEntry, YMD } from "$lib/grpc";
import type { PageData } from "./$types";

export let data: PageData;

let todayContent = data.today.entry?.content || "";

function formatDate(ymd: YMD): string {
	return `${ymd.year}年${ymd.month}月${ymd.day}日`;
}

function getMonthlyUrl(): string {
	const now = new Date();
	return `/monthly/${now.getFullYear()}/${now.getMonth() + 1}`;
}

function formatDateStr(ymd: YMD): string {
	return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
}

function viewEntry(entry: DiaryEntry) {
	const date = entry.date;
	if (date) {
		const dateStr = formatDateStr(date);
		goto(`/${dateStr}`);
	}
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">{$_('diary.title')}</h1>
		<div class="flex gap-3">
			<a
				href={getMonthlyUrl()}
				class="bg-green-600 hover:bg-green-700 text-white font-bold py-2 px-4 rounded"
			>
				{$_('diary.thisMonth')}
			</a>
			<a
				href="/search"
				class="bg-gray-600 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded"
			>
				{$_('diary.search')}
			</a>
		</div>
	</div>

	<div class="space-y-6">
		<!-- 今日の日記 -->
		<div class="bg-white shadow rounded-lg p-6">
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-xl font-semibold text-gray-900">
					{$_('diary.today')} ({formatDate(data.today.date)})
				</h2>
			</div>

			<form method="POST" action="?/saveToday" use:enhance={() => {
				return async ({ result }) => {
					if (result.type === 'success') {
						window.location.reload();
					}
				};
			}}>
				<input type="hidden" name="date" value={formatDateStr(data.today.date)} />
				{#if data.today.entry}
					<input type="hidden" name="id" value={data.today.entry.id} />
				{/if}
				<div class="mb-4">
					<textarea
						name="content"
						bind:value={todayContent}
						placeholder={$_('diary.placeholder')}
						rows="8"
						class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 resize-none"
					></textarea>
				</div>
				<div class="flex justify-end">
					<button
						type="submit"
						class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md font-medium"
					>
						{$_('diary.save')}
					</button>
				</div>
			</form>
		</div>

		<!-- 昨日の日記 -->
		<div class="bg-white shadow rounded-lg p-6">
			<h2 class="text-xl font-semibold text-gray-900 mb-4">
				{$_('diary.yesterday')} ({formatDate(data.yesterday.date)})
			</h2>
			{#if data.yesterday.entry}
				<div class="text-gray-700 whitespace-pre-wrap">
					{data.yesterday.entry.content || ''}
				</div>
				<div class="mt-4">
					<button
						on:click={() => data.yesterday.entry && viewEntry(data.yesterday.entry)}
						class="text-blue-600 hover:text-blue-800 font-medium"
					>
						{$_('diary.viewDetail')}
					</button>
				</div>
			{:else}
				<p class="text-gray-500">{$_('diary.yesterdayNoEntry')}</p>
			{/if}
		</div>

		<!-- 一昨日の日記 -->
		<div class="bg-white shadow rounded-lg p-6">
			<h2 class="text-xl font-semibold text-gray-900 mb-4">
				{$_('diary.dayBeforeYesterday')} ({formatDate(data.dayBeforeYesterday.date)})
			</h2>
			{#if data.dayBeforeYesterday.entry}
				<div class="text-gray-700 whitespace-pre-wrap">
					{data.dayBeforeYesterday.entry.content || ''}
				</div>
				<div class="mt-4">
					<button
						on:click={() => data.dayBeforeYesterday.entry && viewEntry(data.dayBeforeYesterday.entry)}
						class="text-blue-600 hover:text-blue-800 font-medium"
					>
						{$_('diary.viewDetail')}
					</button>
				</div>
			{:else}
				<p class="text-gray-500">{$_('diary.dayBeforeYesterdayNoEntry')}</p>
			{/if}
		</div>
	</div>
</div>