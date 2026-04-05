<script lang="ts">
	import { _ } from "svelte-i18n";
	import { goto } from "$app/navigation";
	import "$lib/i18n";
	import type { DiaryEntry } from "$lib/grpc/diary/diary_pb";
	import type { SemanticSearchResult } from "$lib/grpc/diary/diary_pb";
	import type { PageData } from "./$types";

	type TextSegment = { text: string; isMatch: boolean };

	export let data: PageData;

	let searchKeyword = data.keyword || "";
	// 意味的検索が無効の場合はキーワードモードにフォールバック
	let searchMode: "keyword" | "semantic" =
		data.semanticSearchEnabled && data.mode === "semantic"
			? "semantic"
			: "keyword";

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

	function _viewSemanticEntry(result: SemanticSearchResult) {
		const date = result.date;
		if (date) {
			const dateStr = formatDateUrl(date);
			goto(`/${dateStr}`);
		}
	}

	function _handleSearch() {
		if (searchKeyword.trim()) {
			goto(
				`/search?q=${encodeURIComponent(searchKeyword.trim())}&mode=${searchMode}`,
			);
		}
	}

	function _handleKeydown(event: KeyboardEvent) {
		if (event.key === "Enter") {
			_handleSearch();
		}
	}

	// エントリ一覧の中でキーワードに一致するエントリ数を返す
	function _countKeywordInEntries(
		entries: DiaryEntry[],
		keyword: string,
	): number {
		if (!keyword) return 0;
		const lower = keyword.toLowerCase();
		return entries.filter((e) => e.content.toLowerCase().includes(lower))
			.length;
	}

	// テキストを複数キーワードで分割してセグメント配列を返す
	function _getSegments(text: string, keywords: string[]): TextSegment[] {
		const WINDOW = 150;
		// 改行（CR/LF/CRLF）を空白に置換して1行で表示し、連続スペースも正規化
		text = text
			.replace(/\r\n|\r|\n/g, " ")
			.replace(/ {2,}/g, " ")
			.trim();

		const activeKeywords = keywords.filter((k) => k.trim());
		if (activeKeywords.length === 0) {
			const truncated =
				text.length > WINDOW ? `${text.substring(0, WINDOW)}...` : text;
			return [{ text: truncated, isMatch: false }];
		}

		// 全キーワードから最初のマッチ位置を検索（大文字小文字無視）
		const lowerText = text.toLowerCase();
		let firstMatchIndex = -1;
		for (const kw of activeKeywords) {
			const idx = lowerText.indexOf(kw.trim().toLowerCase());
			if (idx !== -1 && (firstMatchIndex === -1 || idx < firstMatchIndex)) {
				firstMatchIndex = idx;
			}
		}

		let excerpt: string;
		let prefix = "";
		let suffix = "";

		if (firstMatchIndex === -1 || firstMatchIndex < WINDOW) {
			// 1. キーワードが冒頭150文字内に存在する場合（または見つからない場合）は冒頭から表示
			excerpt = text.length > WINDOW ? text.substring(0, WINDOW) : text;
			if (text.length > WINDOW) suffix = "...";
		} else {
			// 2. キーワードが冒頭150文字以降にある場合：ハイライトが中央になるよう前後を切り出す
			const half = Math.floor(WINDOW / 2);
			const start = Math.max(0, firstMatchIndex - half);
			const end = Math.min(text.length, start + WINDOW);
			excerpt = text.substring(start, end);
			if (start > 0) prefix = "...";
			if (end < text.length) suffix = "...";
		}

		// 全キーワードを結合した正規表現でexcerptを分割してセグメント配列を生成
		const escapedKeywords = activeKeywords.map((k) =>
			k.trim().replace(/[.*+?^${}()|[\]\\]/g, "\\$&"),
		);
		const regex = new RegExp(`(${escapedKeywords.join("|")})`, "gi");
		const parts = excerpt.split(regex);

		const lowerKeywords = activeKeywords.map((k) => k.trim().toLowerCase());
		const segments: TextSegment[] = [];
		if (prefix) segments.push({ text: prefix, isMatch: false });
		for (const part of parts) {
			if (part) {
				segments.push({
					text: part,
					isMatch: lowerKeywords.some((kw) => part.toLowerCase() === kw),
				});
			}
		}
		if (suffix) segments.push({ text: suffix, isMatch: false });

		return segments;
	}

	function _formatSimilarity(similarity: number): string {
		return `${Math.round(similarity * 100)}%`;
	}
</script>

<svelte:head>
	<title>{$_('search.title')} - umi.mikan</title>
</svelte:head>

<div class="max-w-4xl mx-auto p-6">
	<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-6">{$_('search.title')}</h1>

	<!-- 検索フォーム -->
	<div class="mb-8">
		<!-- 検索モード切り替えトグル（意味的検索が有効な場合のみ表示） -->
		{#if data.semanticSearchEnabled}
			<div class="flex gap-2 mb-3">
				<button
					on:click={() => { searchMode = 'keyword'; }}
					class="px-4 py-1.5 text-sm rounded-full transition-colors {searchMode === 'keyword'
						? 'bg-blue-600 dark:bg-blue-500 text-white'
						: 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'}"
				>
					{$_('search.modeKeyword')}
				</button>
				<button
					on:click={() => { searchMode = 'semantic'; }}
					class="px-4 py-1.5 text-sm rounded-full transition-colors {searchMode === 'semantic'
						? 'bg-purple-600 dark:bg-purple-500 text-white'
						: 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'}"
				>
					{$_('search.modeSemantic')}
				</button>
			</div>
		{/if}

		<div class="flex gap-4">
			<input
				type="text"
				bind:value={searchKeyword}
				on:keydown={_handleKeydown}
				placeholder={searchMode === 'semantic' ? $_('search.placeholderSemantic') : $_('search.placeholder')}
				class="flex-1 px-4 py-2 border rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:border-transparent {searchMode === 'semantic'
					? 'border-purple-300 dark:border-purple-600 focus:ring-purple-500'
					: 'border-gray-300 dark:border-gray-600 focus:ring-blue-500'}"
			/>
			<button
				on:click={_handleSearch}
				class="px-6 py-2 text-white rounded-lg focus:outline-none focus:ring-2 focus:ring-offset-2 {searchMode === 'semantic'
					? 'bg-purple-600 dark:bg-purple-500 hover:bg-purple-700 dark:hover:bg-purple-600 focus:ring-purple-500'
					: 'bg-blue-600 dark:bg-blue-500 hover:bg-blue-700 dark:hover:bg-blue-600 focus:ring-blue-500'}"
			>
				{$_('search.button')}
			</button>
		</div>

		{#if searchMode === 'semantic'}
			<p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
				{$_('search.semanticDescription')}
			</p>
		{/if}
	</div>

	<!-- エラー表示 -->
	{#if data.error}
		<div class="bg-red-100 dark:bg-red-900/20 border border-red-400 dark:border-red-600 text-red-700 dark:text-red-400 px-4 py-3 rounded mb-4">
			{$_(`search.error.${data.error}`) || data.error}
		</div>
	{/if}

	<!-- キーワード検索結果 -->
	{#if data.searchResults}
		<div class="mb-4">
			<p class="text-gray-600 dark:text-gray-400">
				「{data.searchResults.searchedKeyword}」{$_('search.results')}: {data.searchResults.entries.length}{$_('search.resultCount')}
			</p>
			{#if data.expandedKeywords && data.expandedKeywords.length > 0}
				<div class="text-sm mt-2">
					<div class="flex items-center gap-2 text-gray-700 dark:text-gray-300">
						<span class="font-medium">{data.searchResults.searchedKeyword}</span>
						<span class="text-gray-400 dark:text-gray-500">({_countKeywordInEntries(data.searchResults.entries, data.searchResults.searchedKeyword)}{$_('search.resultCount')})</span>
					</div>
					<div class="ml-2 mt-1">
						{#each data.expandedKeywords as kw, i}
							<div class="relative flex items-center gap-2 py-0.5 pl-4">
								<div class="absolute left-0 w-0.5 bg-gray-200 dark:bg-gray-700 {i === data.expandedKeywords.length - 1 ? 'top-0 h-1/2' : 'top-0 bottom-0'}"></div>
								<div class="absolute left-0 top-1/2 w-4 h-0.5 bg-gray-200 dark:bg-gray-700 -translate-y-px"></div>
								<span class="text-gray-500 dark:text-gray-400">{kw}</span>
								<span class="text-gray-400 dark:text-gray-500">({_countKeywordInEntries(data.searchResults.entries, kw)}{$_('search.resultCount')})</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}
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
							class="text-gray-700 dark:text-gray-300 text-sm auto-phrase-target"
						>
							<p class="line-clamp-3">
								{#each _getSegments(entry.content, [data.searchResults?.searchedKeyword ?? '', ...(data.expandedKeywords ?? [])]) as segment}
									{#if segment.isMatch}
										<mark class="bg-yellow-200 dark:bg-yellow-800 text-gray-900 dark:text-gray-100 rounded px-0.5">{segment.text}</mark>
									{:else}
										{segment.text}
									{/if}
								{/each}
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

	<!-- 意味的検索結果 -->
	{:else if data.semanticResults}
		<div class="mb-4">
			<p class="text-gray-600 dark:text-gray-400">
				「{data.keyword}」{$_('search.semanticResults')}: {data.semanticResults.results.length}{$_('search.resultCount')}
			</p>
			{#if data.semanticResults.embeddingModel || data.semanticResults.chunkModel}
				<p class="text-xs text-gray-400 dark:text-gray-500 mt-1 space-x-3">
					{#if data.semanticResults.embeddingModel}
						<span>{$_('search.embeddingModel')}: {data.semanticResults.embeddingModel}</span>
					{/if}
					{#if data.semanticResults.chunkModel}
						<span>{$_('search.chunkModel')}: {data.semanticResults.chunkModel}</span>
					{/if}
				</p>
			{/if}
		</div>

		{#if data.semanticResults.results.length > 0}
			<div class="grid gap-4">
				{#each data.semanticResults.results as result}
					<div
						class="bg-white dark:bg-gray-800 border border-purple-200 dark:border-purple-800 rounded-lg p-6 shadow-sm dark:shadow-gray-900/20 hover:shadow-md dark:hover:shadow-gray-900/30 transition-shadow cursor-pointer"
						on:click={() => _viewSemanticEntry(result)}
						on:keydown={(e) => e.key === 'Enter' && _viewSemanticEntry(result)}
						role="button"
						tabindex="0"
					>
						<div class="flex justify-between items-start mb-2">
							<h3 class="text-lg font-semibold text-purple-600 dark:text-purple-400">
								{result.date ? _formatDate(result.date) : $_('diary.dateUnknown')}
							</h3>
							<span class="text-xs px-2 py-1 bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded-full">
								{$_('search.similarity')}: {_formatSimilarity(result.similarity)}
							</span>
						</div>
						<div class="text-gray-700 dark:text-gray-300 text-sm whitespace-pre-wrap auto-phrase-target">
							<p class="line-clamp-3">{result.snippet}</p>
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<div class="text-center py-8 text-gray-500 dark:text-gray-400">
				<p>{$_('search.noResults')}</p>
				<p class="text-sm text-gray-500 dark:text-gray-400 mt-2">{$_('search.semanticNoResultsHint')}</p>
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
