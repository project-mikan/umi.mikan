<script lang="ts">
	import { _, locale } from "svelte-i18n";
	import { enhance } from "$app/forms";
	import { goto, invalidateAll } from "$app/navigation";
	import { beforeNavigate } from "$app/navigation";
	import { page } from "$app/stores";
	import { onMount } from "svelte";
	import "$lib/i18n";
	import Button from "$lib/components/atoms/Button.svelte";
	import SaveButton from "$lib/components/atoms/SaveButton.svelte";
	import DiaryCard from "$lib/components/molecules/DiaryCard.svelte";
	import DiaryNavigation from "$lib/components/molecules/DiaryNavigation.svelte";
	import FormField from "$lib/components/molecules/FormField.svelte";
	import HighlightDisplay from "$lib/components/molecules/HighlightDisplay.svelte";
	import Modal from "$lib/components/molecules/Modal.svelte";
	import PastEntriesLinks from "$lib/components/molecules/PastEntriesLinks.svelte";
	import SummaryDisplay from "$lib/components/molecules/SummaryDisplay.svelte";
	import { getDayOfWeekKey } from "$lib/utils/date-utils";
	import { createSubmitHandler } from "$lib/utils/form-utils";
	import type { HighlightData } from "$lib/types/highlight";
	import type { ActionData, PageData } from "./$types";

	export let data: PageData;
	export let form: ActionData;

	// Reactive date formatting function
	$: _formatDate = (ymd: {
		year: number;
		month: number;
		day: number;
	}): string => {
		const dayOfWeekKey = getDayOfWeekKey(ymd);
		const dayOfWeek = $_(`date.dayOfWeek.${dayOfWeekKey}`);
		return $_("date.format.yearMonthDayWithDayOfWeek", {
			values: {
				year: ymd.year,
				month: ymd.month,
				day: ymd.day,
				dayOfWeek: dayOfWeek,
			},
		});
	};

	$: title = $_("page.title.individual", {
		values: {
			date: _formatDate(data.date),
		},
	});

	$: content = data.entry?.content || "";
	let formElement: HTMLFormElement;
	let _showDeleteConfirm = false;
	let loading = false;
	let saved = false;
	let summary: {
		id: string;
		diaryId: string;
		date: { year: number; month: number; day: number };
		summary: string;
		createdAt: number;
		updatedAt: number;
	} | null = data.dailySummary;
	let summaryError: string | null = null;
	let isToday = false;
	let isFutureDate = false;
	let isSummaryGenerating = false; // 要約生成中のフラグ
	let lastSummaryUpdateTime = 0; // 最後に要約が更新された時刻（ミリ秒）
	let isHighlightOutdated = false; // ハイライトが古いかどうか

	// ハイライトデータ
	let highlightData: HighlightData | null = null;
	let highlightVisible = true; // ハイライトの表示・非表示状態
	let embeddingDetailOpen = false; // vectorの詳細表示トグル

	// 検索ハイライト（URLの?searchパラメータから取得）
	$: searchKeyword = $page.url.searchParams.get("search") ?? "";

	// 検索ハイライトをクリア
	function _clearSearchHighlight() {
		goto($page.url.pathname, { replaceState: true });
	}

	// Textareaに渡すハイライトデータ（表示・非表示を反映）
	// 配列の参照を毎回変更してTextareaのリアクティビティを確実にトリガー
	$: displayedHighlights =
		highlightVisible && highlightData ? [...highlightData.highlights] : [];

	// 未保存状態の管理
	let initialContent = "";
	let allowNavigation = false;

	// 前回のdataを保持して変更を検出
	let previousEntryId = "";

	// コンテンツの変更を監視して未保存状態を更新
	$: hasUnsavedChanges = content !== initialContent && !allowNavigation;

	// Check if user has LLM key configured
	$: existingLLMKey = data.user?.llmKeys?.find((key) => key.llmProvider === 1);
	$: hasLLMKey = !!existingLLMKey;

	// 日付判定（当日・未来日）
	$: {
		const now = new Date();
		const currentDate = new Date(
			data.date.year,
			data.date.month - 1,
			data.date.day,
		);
		const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());

		isToday = currentDate.getTime() === today.getTime();
		isFutureDate = currentDate.getTime() > today.getTime();
	}
	$: autoSummaryDisabled = !existingLLMKey?.autoSummaryDaily;

	// 無効化メッセージを取得
	function getDisabledMessage(): string {
		if (isFutureDate) {
			return $_("diary.summaryNotAvailableFuture");
		}
		if (isToday) {
			return $_("diary.summaryNotAvailableToday");
		}
		return "";
	}

	// Check if the diary date is not today (only allow summary generation for past entries)
	$: isNotToday = (() => {
		if (!data.today) return false;

		return (
			data.date.year < data.today.year ||
			(data.date.year === data.today.year &&
				data.date.month < data.today.month) ||
			(data.date.year === data.today.year &&
				data.date.month === data.today.month &&
				data.date.day < data.today.day)
		);
	})();

	// Check if summary is outdated (diary updatedAt > summary updatedAt)
	$: isSummaryOutdated = (() => {
		if (!summary || !data.entry) return false;

		// 日記エントリは秒単位、サマリーはミリ秒単位なので統一
		const diaryUpdatedAt = Number(data.entry.updatedAt) * 1000; // 秒 → ミリ秒
		const summaryUpdatedAt = Number(summary.updatedAt); // 既にミリ秒

		// 要約が最近更新された場合（5秒以内）は古くないとみなす
		const now = Date.now();
		const recentlyUpdated =
			lastSummaryUpdateTime > 0 && now - lastSummaryUpdateTime < 5000;

		// 要約が日記よりも新しい場合、または最近更新された場合は古くない
		const isOutdated = diaryUpdatedAt > summaryUpdatedAt && !recentlyUpdated;

		return isOutdated;
	})();

	// Check if highlight is outdated (diary updatedAt > highlight updatedAt)
	$: {
		if (!highlightData || !data.entry) {
			isHighlightOutdated = false;
		} else {
			// 日記エントリは秒単位、ハイライトはミリ秒単位なので統一
			const diaryUpdatedAt = Number(data.entry.updatedAt) * 1000; // 秒 → ミリ秒
			const highlightUpdatedAt = Number(highlightData.updatedAt); // 既にミリ秒

			// ハイライトが日記よりも古い場合
			const isOutdated = diaryUpdatedAt > highlightUpdatedAt;

			// 古くなった場合は非表示にする
			if (isOutdated && !isHighlightOutdated) {
				isHighlightOutdated = true;
				highlightVisible = false;
			} else if (!isOutdated) {
				isHighlightOutdated = false;
			}
		}
	}

	// Character count calculation
	$: characterCount = content ? content.length : 0;

	// データが変更された時に要約状態を更新
	// ページ遷移時のみ（entryのIDが変わった時のみ）実行
	$: {
		// entryの一意性を判定するためのID
		const currentEntryId =
			data.entry?.id || `${data.date.year}-${data.date.month}-${data.date.day}`;

		// ページが変更された場合のみ初期化
		if (currentEntryId !== previousEntryId) {
			previousEntryId = currentEntryId;

			// 要約とコンテンツを更新
			summary = data.dailySummary;
			isSummaryGenerating = false;

			// 初期コンテンツを設定
			initialContent = data.entry?.content || "";

			// コンテンツ変数を初期化（ユーザー入力を上書きしない）
			if (content !== initialContent) {
				content = initialContent;
			}

			// 新しいページではallowNavigationをリセット
			allowNavigation = false;
		}
	}

	function handleSummaryUpdated(event: CustomEvent) {
		const newSummary = event.detail.summary;
		const oldSummary = summary;

		// 要約が実際に変更されたかどうかを確認
		const actuallyUpdated =
			!oldSummary ||
			oldSummary.updatedAt !== newSummary.updatedAt ||
			oldSummary.summary !== newSummary.summary;

		summary = newSummary;

		// 要約が実際に更新された場合のみ時刻を記録
		if (actuallyUpdated) {
			lastSummaryUpdateTime = Date.now();
		}
	}

	function handleSummaryError(event: CustomEvent) {
		summaryError = event.detail.message;
	}

	function handleGenerationStarted() {
		isSummaryGenerating = true;
		summaryError = null;
	}

	function handleGenerationCompleted() {
		isSummaryGenerating = false;
	}

	function handleHighlightUpdated(event: CustomEvent) {
		const newHighlight = event.detail.highlight;
		highlightData = newHighlight;
		// ハイライトが更新されたら古くないとマークして表示する
		isHighlightOutdated = false;
		highlightVisible = true;
	}

	function handleHighlightVisibilityChanged(event: CustomEvent) {
		// ハイライトの表示・非表示状態を更新
		highlightVisible = event.detail.visible;
	}

	function _formatDateStr(ymd: {
		year: number;
		month: number;
		day: number;
	}): string {
		return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
	}

	function _goBack() {
		goto("/");
	}

	function _goToMonthly() {
		const year = data.date.year;
		const month = String(data.date.month).padStart(2, "0");
		goto(`/monthly/${year}/${month}`);
	}

	function _handleSave() {
		formElement?.requestSubmit();
	}

	function _confirmDelete() {
		_showDeleteConfirm = true;
	}

	function _cancelDelete() {
		_showDeleteConfirm = false;
	}

	function _handleDelete() {
		// 削除時は遷移を許可
		allowNavigation = true;
		const form = document.createElement("form");
		form.method = "POST";
		form.action = "?/delete";
		document.body.appendChild(form);
		form.submit();
	}

	// ページ遷移前の警告
	beforeNavigate((navigation) => {
		if (hasUnsavedChanges && !allowNavigation) {
			if (!confirm($_("diary.unsavedChangesWarning"))) {
				navigation.cancel();
			}
		}
	});

	// モバイルキーボード表示時の保存ボタン位置調整
	let saveButtonBottom = "5rem"; // デフォルト: bottom-20 (QuickNavigationの上)

	// ブラウザのページ離脱時の警告
	onMount(() => {
		const handleBeforeUnload = (e: BeforeUnloadEvent) => {
			if (hasUnsavedChanges) {
				e.preventDefault();
				e.returnValue = "";
			}
		};

		// visualViewport APIを使ってキーボード表示を検出し、保存ボタンの位置を調整
		const updateSaveButtonPosition = () => {
			if (!window.visualViewport) return;
			const viewport = window.visualViewport;
			const keyboardHeight = window.innerHeight - viewport.height;
			const KEYBOARD_THRESHOLD = 100;
			if (keyboardHeight > KEYBOARD_THRESHOLD) {
				saveButtonBottom = `${keyboardHeight + 8}px`;
			} else {
				saveButtonBottom = "5rem";
			}
		};

		if (window.visualViewport) {
			window.visualViewport.addEventListener(
				"resize",
				updateSaveButtonPosition,
			);
			window.visualViewport.addEventListener(
				"scroll",
				updateSaveButtonPosition,
			);
		}

		window.addEventListener("beforeunload", handleBeforeUnload);

		return () => {
			window.removeEventListener("beforeunload", handleBeforeUnload);
			if (window.visualViewport) {
				window.visualViewport.removeEventListener(
					"resize",
					updateSaveButtonPosition,
				);
				window.visualViewport.removeEventListener(
					"scroll",
					updateSaveButtonPosition,
				);
			}
		};
	});
</script>

<svelte:head>
	<title>{title}</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">{$_("diary.title")}</h1>
		<div class="flex gap-2">
			<button
				on:click={_goToMonthly}
				class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium"
			>
				{$_("diary.viewThisMonth")}
			</button>
		</div>
	</div>

	<div class="space-y-6">
		<DiaryNavigation currentDate={data.date} />

		<!-- Summary display area -->
		{#if data.entry && characterCount >= 1000}
			<SummaryDisplay
				type="daily"
				{summary}
				fetchUrl="/api/diary/summary/daily/{data.date.year}/{data.date.month}/{data.date.day}"
				generateUrl="/api/diary/summary/generate-daily"
				generatePayload={{
					diaryId: data.entry.id,
					content: data.entry.content,
					date: data.date
				}}
				{hasLLMKey}
				{isSummaryOutdated}
				isDisabled={isToday || isFutureDate}
				disabledMessage={getDisabledMessage()}
				isGenerating={isSummaryGenerating}
				on:summaryUpdated={handleSummaryUpdated}
				on:error={handleSummaryError}
				on:generationStarted={handleGenerationStarted}
				on:generationCompleted={handleGenerationCompleted}
			/>
		{/if}

		<!-- Summary error display -->
		{#if summaryError}
			<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 auto-phrase-target">
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
						</svg>
					</div>
					<div class="ml-3">
						<p class="text-sm text-red-800 dark:text-red-200">{summaryError}</p>
					</div>
				</div>
			</div>
		{/if}

		<DiaryCard
			title=""
			entry={data.entry}
			showForm={true}
		>
			<form
				bind:this={formElement}
				method="POST"
				action="?/save"
use:enhance={createSubmitHandler(
	(l) => loading = l,
	async (s) => {
		saved = s;
		if (s) {
			// 保存成功時にページデータを再読み込み（updatedAtを更新）
			await invalidateAll();
			// 保存成功時に初期コンテンツを更新
			initialContent = content;
			// hasUnsavedChangesの再計算に任せる
			// ハイライトの古い/新しいの判定はリアクティブステートメントに任せる
		}
	}
)}
				slot="form"
			>
				<input type="hidden" name="date" value={_formatDateStr(data.date)} />
				{#if data.entry}
					<input type="hidden" name="id" value={data.entry.id} />
				{/if}
				<!-- Highlight controls -->
				{#if data.entry && characterCount >= 500}
					<HighlightDisplay
						diaryId={data.entry.id}
						{hasLLMKey}
						{isHighlightOutdated}
						diaryUpdatedAt={Number(data.entry.updatedAt)}
						on:highlightUpdated={handleHighlightUpdated}
						on:highlightVisibilityChanged={handleHighlightVisibilityChanged}
					/>
				{/if}

				<!-- 検索ハイライトバナー -->
				{#if searchKeyword}
					<div class="flex items-center gap-2 px-3 py-2 mb-2 bg-orange-50 dark:bg-orange-900/20 border border-orange-200 dark:border-orange-700 rounded-md text-sm">
						<svg class="w-4 h-4 text-orange-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
							<path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
						</svg>
						<span class="text-orange-800 dark:text-orange-200 flex-1">
							{$_("diary.searchHighlightBanner", { values: { keyword: searchKeyword } })}
						</span>
						<button
							type="button"
							on:click={_clearSearchHighlight}
							class="text-orange-600 dark:text-orange-400 hover:text-orange-800 dark:hover:text-orange-200 font-medium"
						>
							{$_("diary.searchHighlightClear")}
						</button>
					</div>
				{/if}

				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					diaryHighlights={displayedHighlights}
					{searchKeyword}
					bind:value={content}
					on:save={_handleSave}
					on:autosave={() => { if (hasUnsavedChanges && !loading) _handleSave(); }}
				/>

				{#if form?.error}
					<div class="mt-2 text-sm text-red-600 dark:text-red-400">
						{form.error}
					</div>
				{/if}

				<!-- Character count display -->
				<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{$_("diary.characterCount", { values: { count: characterCount } })}
					{#if characterCount >= 1000}
						<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
							({$_("diary.autoSummaryEligible")})
						</span>
					{/if}
					{#if data.entry?.updatedAt}
						<span class="ml-4">
							{$_("common.updatedAt")}: {new Date(Number(data.entry.updatedAt) * 1000).toLocaleString()}
						</span>
					{/if}
				</div>

				<!-- RAGインデックス状態 -->
				{#if data.semanticSearchEnabled && data.entry}
					<div class="mt-2 text-xs">
						<div class="flex items-center gap-2">
							{#if data.embeddingStatus?.indexed}
								<button
									type="button"
									class="inline-flex items-center px-2 py-0.5 rounded-full bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300 hover:bg-purple-200 dark:hover:bg-purple-900/50 transition-colors cursor-pointer"
									on:click={() => (embeddingDetailOpen = !embeddingDetailOpen)}
								>
									<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
										<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
									</svg>
									{$_("diary.embedding.indexed")}
									<svg class="w-3 h-3 ml-1 transition-transform {embeddingDetailOpen ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
									</svg>
								</button>
							{:else}
								<span class="inline-flex items-center px-2 py-0.5 rounded-full bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400">
									<svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
										<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
									</svg>
									{$_("diary.embedding.notIndexed")}
								</span>
							{/if}
						</div>

						{#if data.embeddingStatus?.indexed && embeddingDetailOpen}
							<div class="mt-2 p-3 rounded-lg bg-purple-50 dark:bg-purple-950/20 border border-purple-200 dark:border-purple-800/40 space-y-1.5">
								<div class="flex items-center gap-2">
									<span class="text-gray-500 dark:text-gray-400 w-28 shrink-0">{$_("diary.embedding.modelVersion")}:</span>
									<span class="font-mono text-purple-700 dark:text-purple-300">{data.embeddingStatus.modelVersion}</span>
								</div>
								<div class="flex items-center gap-2">
									<span class="text-gray-500 dark:text-gray-400 w-28 shrink-0">{$_("diary.embedding.chunkCount")}:</span>
									<span class="text-gray-700 dark:text-gray-300">{data.embeddingStatus.chunkCount}</span>
								</div>
								{#if data.embeddingStatus.chunkSummaries.length > 0}
									<div>
										<span class="text-gray-500 dark:text-gray-400">{$_("diary.embedding.chunkSummaries")}:</span>
										<ol class="mt-1 space-y-0.5 list-decimal list-inside">
											{#each data.embeddingStatus.chunkSummaries as chunkSummary}
												<li class="text-gray-600 dark:text-gray-400">
													{#if chunkSummary}
														{chunkSummary}
													{:else}
														{$_("diary.embedding.chunkSummaryEmpty")}
													{/if}
												</li>
											{/each}
										</ol>
									</div>
								{/if}
								<div class="flex items-center gap-2">
									<span class="text-gray-500 dark:text-gray-400 w-28 shrink-0">{$_("diary.embedding.dimensions")}:</span>
									<span class="text-gray-700 dark:text-gray-300">{data.embeddingStatus.embeddingDimensions}</span>
								</div>
								<div class="flex items-center gap-2">
									<span class="text-gray-500 dark:text-gray-400 w-28 shrink-0">{$_("diary.embedding.indexedAt")}:</span>
									<span class="text-gray-700 dark:text-gray-300">{new Date(data.embeddingStatus.createdAt * 1000).toLocaleString()}</span>
								</div>
								<div class="flex items-center gap-2">
									<span class="text-gray-500 dark:text-gray-400 w-28 shrink-0">{$_("diary.embedding.updatedAt")}:</span>
									<span class="text-gray-700 dark:text-gray-300">{new Date(data.embeddingStatus.updatedAt * 1000).toLocaleString()}</span>
								</div>
								{#if data.embeddingStatus.embeddingDimensions > 0}
									<div>
										<span class="text-gray-500 dark:text-gray-400">{$_("diary.embedding.vectorPreview")}:</span>
										<div class="mt-1 font-mono text-xs text-gray-600 dark:text-gray-400 bg-white dark:bg-gray-900 rounded p-2 border border-purple-100 dark:border-purple-900/40 overflow-x-auto whitespace-nowrap">
											[{data.embeddingStatus.embeddingValues.slice(0, 10).map((v) => v.toFixed(6)).join(", ")}, ...]
										</div>
									</div>
								{/if}
							</div>
						{/if}
					</div>
				{/if}


				<div class="flex justify-between">
					<div>
						{#if data.entry}
							<Button
								type="button"
								variant="danger"
								size="md"
								on:click={_confirmDelete}
							>
								{$_("diary.delete")}
							</Button>
						{/if}
					</div>
				</div>

				<div class="sticky bottom-4 flex justify-end hidden sm:flex mt-4 z-10">
					<SaveButton {loading} {saved} />
				</div>
			</form>
		</DiaryCard>

		<PastEntriesLinks pastEntries={data.pastEntries} />
	</div>

	<!-- Fixed Save Button for Mobile -->
	<div class="fixed left-0 right-0 p-4 sm:hidden z-10 pointer-events-none transition-[bottom] duration-150" style="bottom: {saveButtonBottom}">
		<div class="max-w-4xl mx-auto flex justify-end pointer-events-auto">
			<SaveButton
				type="button"
				{loading}
				{saved}
				size="md"
				on:click={_handleSave}
			/>
		</div>
	</div>

	<!-- Spacer for fixed button on mobile -->
	<div class="h-32 sm:hidden"></div>
</div>

<Modal
	isOpen={_showDeleteConfirm}
	title={$_("edit.deleteConfirm")}
	confirmText={$_("diary.delete")}
	cancelText={$_("diary.cancel")}
	variant="danger"
	onConfirm={_handleDelete}
	onCancel={_cancelDelete}
>
	<p class="text-sm text-gray-500 dark:text-gray-400">
		{$_("edit.deleteMessage")}
	</p>
</Modal>


