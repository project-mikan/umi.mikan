<script lang="ts">
import { _ } from "svelte-i18n";
import { goto } from "$app/navigation";
import "$lib/i18n";
import type { DiaryEntry } from "$lib/grpc/diary/diary_pb";
import { autoPhraseEnabled } from "$lib/auto-phrase-store";
import type { PageData } from "./$types";

export let data: PageData;

let searchKeyword = data.keyword || "";

function _formatDate(ymd: {
	year: number;
	month: number;
	day: number;
}): string {
	return `${ymd.year}年${ymd.month}月${ymd.day}日`;
}

function formatDateUrl(ymd: {
	year: number;
	month: number;
	day: number;
}): string {
	return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
}

function _viewEntry(entry: DiaryEntry) {
	const date = entry.date;
	if (date) {
		const dateStr = formatDateUrl(date);
		goto(`/${dateStr}`);
	}
}

function _handleSearch() {
	if (searchKeyword.trim()) {
		goto(`/search?q=${encodeURIComponent(searchKeyword.trim())}`);
	}
}

function _handleKeydown(event: KeyboardEvent) {
	if (event.key === "Enter") {
		_handleSearch();
	}
}
</script>

<svelte:head>
	<title>{$_('search.title')} - umi.mikan</title>
</svelte:head>

<div class="max-w-4xl mx-auto p-6">
	<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-6">{$_('search.title')}</h1>
	
	<!-- 検索フォーム -->
	<div class="mb-8">
		<div class="flex gap-4">
			<input
				type="text"
				bind:value={searchKeyword}
				on:keydown={_handleKeydown}
				placeholder={$_('search.placeholder')}
				class="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
			/>
			<button
				on:click={_handleSearch}
				class="px-6 py-2 bg-blue-600 dark:bg-blue-500 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
			>
				{$_('search.button')}
			</button>
		</div>
	</div>

	<!-- 検索結果 -->
	{#if data.error}
		<div class="bg-red-100 dark:bg-red-900/20 border border-red-400 dark:border-red-600 text-red-700 dark:text-red-400 px-4 py-3 rounded mb-4">
			{data.error}
		</div>
	{/if}

	{#if data.searchResults}
		<div class="mb-4">
			<p class="text-gray-600 dark:text-gray-400">
				「{data.searchResults.searchedKeyword}」{$_('search.results')}: {data.searchResults.entries.length}{$_('search.resultCount')}
			</p>
		</div>

		{#if data.searchResults.entries.length > 0}
			<div class="grid gap-4">
				{#each data.searchResults.entries as entry}
					<div 
						class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 shadow-sm dark:shadow-gray-900/20 hover:shadow-md dark:hover:shadow-gray-900/30 transition-shadow cursor-pointer"
						on:click={() => _viewEntry(entry)}
						on:keydown={(e) => e.key === 'Enter' && _viewEntry(entry)}
						role="button"
						tabindex="0"
					>
						<div class="flex justify-between items-start mb-2">
							<h3 class="text-lg font-semibold text-blue-600 dark:text-blue-400">
								{entry.date ? _formatDate(entry.date) : $_('diary.dateUnknown')}
							</h3>
						</div>
						<div
							class="text-gray-700 dark:text-gray-300 text-sm whitespace-pre-wrap {$autoPhraseEnabled ? 'auto-phrase' : ''}"
						>
							<p class="line-clamp-3">
								{entry.content.length > 150
									? entry.content.substring(0, 150) + '...'
									: entry.content}
							</p>
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<div class="text-center py-8 text-gray-500 dark:text-gray-400">
				<p>{$_('search.noResults')}</p>
				<p class="text-sm text-gray-500 dark:text-gray-400 mt-2">{$_('search.noResultsHint')}</p>
			</div>
		{/if}
	{:else if data.keyword}
		<div class="text-center py-8 text-gray-500 dark:text-gray-400">
			<p>{$_('search.searching')}</p>
		</div>
	{:else}
		<div class="text-center py-8 text-gray-500 dark:text-gray-400">
			<p>{$_('search.enterKeyword')}</p>
		</div>
	{/if}

</div>

<style>
	.line-clamp-3 {
		display: -webkit-box;
		-webkit-line-clamp: 3;
		line-clamp: 3;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>