<script lang="ts">
	import { _, locale } from "svelte-i18n";
	import { enhance } from "$app/forms";
	import { goto } from "$app/navigation";
	import { beforeNavigate } from "$app/navigation";
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
	let selectedEntities: {
		entityId: string;
		positions: { start: number; end: number }[];
	}[] = [];
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
	interface Highlight {
		start: number;
		end: number;
		text: string;
	}
	interface HighlightData {
		highlights: Highlight[];
		createdAt: number;
		updatedAt: number;
	}
	let highlightData: HighlightData | null = null;
	let highlightVisible = true; // ハイライトの表示・非表示状態

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
		// ハイライトが更新されたら古くないとマーク
		isHighlightOutdated = false;
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

	// ブラウザのページ離脱時の警告
	onMount(() => {
		const handleBeforeUnload = (e: BeforeUnloadEvent) => {
			if (hasUnsavedChanges) {
				e.preventDefault();
				e.returnValue = "";
			}
		};

		window.addEventListener("beforeunload", handleBeforeUnload);

		return () => {
			window.removeEventListener("beforeunload", handleBeforeUnload);
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
	(s) => {
		saved = s;
		if (s) {
			// 保存成功時に初期コンテンツを更新
			initialContent = content;
			// hasUnsavedChangesの再計算に任せる
		}
	}
)}
				slot="form"
			>
				<input type="hidden" name="date" value={_formatDateStr(data.date)} />
				{#if data.entry}
					<input type="hidden" name="id" value={data.entry.id} />
				{/if}
				<input type="hidden" name="selectedEntities" value={JSON.stringify(selectedEntities)} />

				<!-- Highlight controls -->
				{#if data.entry && characterCount >= 100}
					<HighlightDisplay
						diaryId={data.entry.id}
						{hasLLMKey}
						{isHighlightOutdated}
						diaryUpdatedAt={Number(data.entry.updatedAt)}
						on:highlightUpdated={handleHighlightUpdated}
						on:highlightVisibilityChanged={handleHighlightVisibilityChanged}
					/>
				{/if}

				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					diaryEntities={data.entry?.diaryEntities || []}
					diaryHighlights={displayedHighlights}
					bind:value={content}
					bind:selectedEntities
					on:save={_handleSave}
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
	<div class="fixed bottom-20 left-0 right-0 p-4 sm:hidden z-10 pointer-events-none">
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


