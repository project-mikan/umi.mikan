<script lang="ts">
import { _, locale } from "svelte-i18n";
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import "$lib/i18n";
import Button from "$lib/components/atoms/Button.svelte";
import SaveButton from "$lib/components/atoms/SaveButton.svelte";
import DiaryCard from "$lib/components/molecules/DiaryCard.svelte";
import DiaryNavigation from "$lib/components/molecules/DiaryNavigation.svelte";
import FormField from "$lib/components/molecules/FormField.svelte";
import Modal from "$lib/components/molecules/Modal.svelte";
import PastEntriesLinks from "$lib/components/molecules/PastEntriesLinks.svelte";
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
let summaryGenerating = false;
let summaryError: string | null = null;
let summaryStatus: "none" | "queued" | "processing" | "completed" | "error" =
	data.dailySummary ? "completed" : "none";
let summary: {
	id: string;
	diaryId: string;
	date: { year: number; month: number; day: number };
	summary: string;
	createdAt: number;
	updatedAt: number;
} | null = data.dailySummary;
let showSummary = !!data.dailySummary;
let pollingInterval: number | null = null;
let summaryJustUpdated = false;

// Check if user has LLM key configured
$: existingLLMKey = data.user?.llmKeys?.find((key) => key.llmProvider === 1);
$: hasLLMKey = !!existingLLMKey;
$: autoSummaryDisabled = !existingLLMKey?.autoSummaryDaily;

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

	return diaryUpdatedAt > summaryUpdatedAt;
})();

// Character count calculation
$: characterCount = content ? content.length : 0;

// データが変更された時に要約状態を更新
$: {
	summary = data.dailySummary;
	showSummary = !!data.dailySummary;
	summaryStatus = data.dailySummary ? "completed" : "none";
}

// ポーリング処理のクリーンアップ
function clearPolling() {
	if (pollingInterval) {
		clearInterval(pollingInterval);
		pollingInterval = null;
	}
}

// 要約ステータスをポーリングして取得
async function pollSummaryStatus() {
	if (!data.entry) return;

	try {
		const response = await fetch(
			`/api/diary/summary/daily/${data.date.year}/${data.date.month}/${data.date.day}`,
			{
				method: "GET",
				headers: {
					"Content-Type": "application/json",
				},
			},
		);

		if (response.ok) {
			const result = await response.json();
			// 新しい要約が取得できた場合
			if (
				result.summary &&
				(!summary || result.summary.updatedAt > summary.updatedAt)
			) {
				const isUpdate = !!summary; // 既存要約がある場合は更新
				summary = result.summary;
				summaryStatus = "completed";
				summaryGenerating = false;
				showSummary = true;
				if (isUpdate) {
					triggerSummaryUpdateAnimation();
				}
				clearPolling();
			}
		} else if (response.status === 404) {
			// 要約がまだ存在しない場合は継続してポーリング
		} else {
			// その他のエラーの場合はポーリングを停止
			console.error("Error polling summary status:", response.status);
			clearPolling();
			summaryStatus = "error";
			summaryGenerating = false;
		}
	} catch (error) {
		console.error("Polling error:", error);
	}
}

// コンポーネントがアンマウントされる際にポーリングをクリア
import { onDestroy } from "svelte";
onDestroy(() => {
	clearPolling();
});

// 要約更新時のアニメーション
function triggerSummaryUpdateAnimation() {
	summaryJustUpdated = true;
	setTimeout(() => {
		summaryJustUpdated = false;
	}, 1500); // 1.5秒間アニメーション
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
	const form = document.createElement("form");
	form.method = "POST";
	form.action = "?/delete";
	document.body.appendChild(form);
	form.submit();
}

async function _generateSummary() {
	if (
		!data.entry?.content ||
		summaryGenerating ||
		summaryStatus === "queued" ||
		summaryStatus === "processing"
	) {
		return;
	}
	summaryGenerating = true;
	summaryError = null;
	summaryStatus = "queued";

	// UIを最低500ms間は生成中表示にする
	const minDisplayTime = new Promise((resolve) => setTimeout(resolve, 500));

	try {
		const response = await fetch("/api/diary/summary/generate-daily", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({
				diaryId: data.entry.id,
				content: data.entry.content,
				date: data.date,
			}),
		});

		if (!response.ok) {
			const errorData = await response.json().catch(() => ({}));
			throw new Error(errorData.message || "要約の生成に失敗しました");
		}

		const result = await response.json();

		// 生成がキューに入った場合、ポーリングを開始
		if (
			result.summary?.summary &&
			(result.summary.summary.includes("queued") ||
				result.summary.summary.includes("processing") ||
				result.summary.summary.includes("Updating..."))
		) {
			summaryStatus = result.summary.summary.includes("queued")
				? "queued"
				: result.summary.summary.includes("Updating...")
					? "processing"
					: "processing";
			summaryGenerating = false; // ボタンを有効化
			// 3秒間隔でポーリング開始
			pollingInterval = setInterval(pollSummaryStatus, 3000);
			// 最長2分間ポーリング
			setTimeout(() => {
				if (pollingInterval) {
					clearPolling();
					if (summaryStatus !== "completed") {
						summaryStatus = "error";
						summaryError = $_("diary.summaryTimeout");
					}
				}
			}, 120000);
		} else if (result.summary) {
			// 即座に完了した場合でも最小表示時間を待つ
			await minDisplayTime;
			const isUpdate = !!summary; // 既存要約がある場合は更新
			summary = result.summary;
			summaryStatus = "completed";
			summaryGenerating = false;
			showSummary = true;

			// 要約が更新された場合はフラッシュエフェクト
			if (isUpdate) {
				triggerSummaryUpdateAnimation();
			}
		}
	} catch (error) {
		console.error("Summary generation failed:", error);
		await minDisplayTime;
		summaryStatus = "error";
		summaryError =
			error instanceof Error
				? error.message
				: $_("diary.summaryGenerationFailed");
		summaryGenerating = false;
		// エラーメッセージを5秒後に自動クリア
		setTimeout(() => {
			summaryError = null;
			summaryStatus = summary ? "completed" : "none";
		}, 5000);
	}
}

function _clearSummary() {
	summary = null;
	showSummary = false;
}
</script>

<svelte:head>
	<title>{title}</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">{$_("diary.title")}</h1>
		<div class="flex gap-2">
			{#if summary}
				<button
					on:click={() => showSummary = !showSummary}
					class="px-4 py-2 {showSummary ? 'bg-gray-600 hover:bg-gray-700' : 'bg-blue-600 hover:bg-blue-700'} text-white rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
				>
					{showSummary ? $_("diary.summary.hide") : $_("diary.summary.view")}
				</button>
			{/if}
			{#if data.entry && hasLLMKey}
				{@const isDisabled = summaryGenerating || summaryStatus === 'queued' || summaryStatus === 'processing' || characterCount < 1000 || !isNotToday}
				{@const tooltipMessage =
					summaryError ? summaryError :
					summaryStatus === 'queued' ? $_('diary.summary.statusQueued') :
					summaryStatus === 'processing' ? $_('diary.summary.statusProcessing') :
					!isNotToday ? $_("diary.summaryNotAvailableToday") :
					(characterCount < 1000 ? $_("diary.summaryRequires1000Chars") : "")}
				<div class="relative group">
					<button
						on:click={_generateSummary}
						disabled={isDisabled}
						class="px-4 py-2 {summaryError ? 'bg-red-500 hover:bg-red-600' : (!isDisabled ? 'bg-green-600 hover:bg-green-700' : 'bg-gray-400 cursor-not-allowed')} disabled:bg-gray-400 text-white rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 flex items-center gap-2"
					>
						{#if summaryGenerating || summaryStatus === 'queued' || summaryStatus === 'processing'}
							<svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
							</svg>
						{/if}
						{summaryGenerating || summaryStatus === 'queued' || summaryStatus === 'processing' ?
							(summary ? $_("diary.regeneratingSummary") : $_("diary.generatingSummary")) :
							(summary ? $_("diary.regenerateSummary") : $_("diary.generateSummary"))
						}
					</button>
					{#if (isDisabled && tooltipMessage) || summaryError}
						<div class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-3 py-2 text-sm text-white bg-gray-800 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none whitespace-nowrap z-10">
							{tooltipMessage}
							<div class="absolute top-full left-1/2 transform -translate-x-1/2 w-0 h-0 border-l-4 border-r-4 border-t-4 border-transparent border-t-gray-800"></div>
						</div>
					{/if}
				</div>
			{/if}
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
		{#if (showSummary && summary) || summaryStatus === 'queued' || summaryStatus === 'processing'}
			<div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
				<div class="p-6">
					<!-- Summary status indicator -->
					{#if summaryStatus === 'queued' || summaryStatus === 'processing'}
						<div class="mb-4 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md">
							<div class="flex items-center">
								<svg class="animate-spin h-5 w-5 text-blue-400 mr-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								<p class="text-sm text-blue-800 dark:text-blue-200">
									{summaryStatus === 'queued' ? $_('diary.summary.statusQueued') : $_('diary.summary.statusProcessing')}
								</p>
							</div>
						</div>
					{/if}
					<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
						{$_("diary.summary.label")}
					</h2>
					{#if summary}
						{#if isSummaryOutdated}
							<div class="mb-4 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md">
								<div class="flex">
									<div class="flex-shrink-0">
										<svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
											<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
										</svg>
									</div>
									<div class="ml-3">
										<p class="text-sm text-yellow-800 dark:text-yellow-200">
											{$_("diary.summary.outdatedWarning")}
										</p>
									</div>
								</div>
							</div>
						{/if}
						<div class="prose dark:prose-invert max-w-none">
							<p class="text-gray-700 dark:text-gray-300 whitespace-pre-wrap leading-relaxed transition-all duration-300 px-2 py-1 rounded"
							   class:summary-flash={summaryJustUpdated}>
								{summary.summary}
							</p>
						</div>
					{/if}
					{#if summary}
						<div class="mt-6 flex justify-between items-center text-sm text-gray-500 dark:text-gray-400">
							<span>
								{$_("common.createdAt")}: {new Date(summary.createdAt).toLocaleString()}
							</span>
							<span>
								{$_("common.updatedAt")}: {new Date(summary.updatedAt).toLocaleString()}
							</span>
						</div>
					{/if}
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
use:enhance={createSubmitHandler((l) => loading = l, (s) => saved = s)}
				slot="form"
			>
				<input type="hidden" name="date" value={_formatDateStr(data.date)} />
				{#if data.entry}
					<input type="hidden" name="id" value={data.entry.id} />
				{/if}
				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					bind:value={content}
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
					<SaveButton {loading} {saved} />
				</div>
			</form>
		</DiaryCard>

		<PastEntriesLinks pastEntries={data.pastEntries} />
	</div>
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

<style>
	@keyframes summary-glow {
		0% {
			background-color: transparent;
			box-shadow: inset 0 0 0 1px transparent;
		}
		25% {
			background-color: rgba(59, 130, 246, 0.1);
			box-shadow: inset 0 0 0 1px rgba(59, 130, 246, 0.3);
		}
		50% {
			background-color: rgba(59, 130, 246, 0.15);
			box-shadow: inset 0 0 0 2px rgba(59, 130, 246, 0.5);
		}
		75% {
			background-color: rgba(59, 130, 246, 0.1);
			box-shadow: inset 0 0 0 1px rgba(59, 130, 246, 0.3);
		}
		100% {
			background-color: transparent;
			box-shadow: inset 0 0 0 1px transparent;
		}
	}

	@keyframes summary-glow-dark {
		0% {
			background-color: transparent;
			box-shadow: inset 0 0 0 1px transparent;
		}
		25% {
			background-color: rgba(59, 130, 246, 0.2);
			box-shadow: inset 0 0 0 1px rgba(59, 130, 246, 0.4);
		}
		50% {
			background-color: rgba(59, 130, 246, 0.25);
			box-shadow: inset 0 0 0 2px rgba(59, 130, 246, 0.6);
		}
		75% {
			background-color: rgba(59, 130, 246, 0.2);
			box-shadow: inset 0 0 0 1px rgba(59, 130, 246, 0.4);
		}
		100% {
			background-color: transparent;
			box-shadow: inset 0 0 0 1px transparent;
		}
	}

	:global(.summary-flash) {
		animation: summary-glow 1.5s ease-in-out;
	}

	:global(.dark .summary-flash) {
		animation: summary-glow-dark 1.5s ease-in-out;
	}
</style>

