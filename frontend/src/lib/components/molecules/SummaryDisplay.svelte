<script lang="ts">
import { _, locale } from "svelte-i18n";
import { browser } from "$app/environment";
import { onDestroy, createEventDispatcher } from "svelte";
import { authenticatedFetch } from "$lib/auth-client";
import "$lib/i18n";

export let type: "daily" | "monthly"; // 日次または月次
export let summary: any = null;
export let fetchUrl: string; // 要約取得用URL
export let generateUrl: string; // 要約生成用URL
export let generatePayload: any = {}; // 生成時に送信するペイロード
export let isDisabled = false; // 生成ボタンを無効にするかどうか
export let disabledMessage = ""; // 無効時のメッセージ
export let showSummary = true;
export let hasLLMKey = true;

const dispatch = createEventDispatcher();

let summaryStatus: "none" | "queued" | "processing" | "completed" = summary
	? "completed"
	: "none";
let summaryGenerating = false;
let pollingInterval: ReturnType<typeof setInterval> | null = null;
let summaryJustUpdated = false;

// ポーリング機能
async function pollSummaryStatus(isUpdate = false) {
	if (!browser) return;
	try {
		const response = await authenticatedFetch(fetchUrl);
		if (response.ok) {
			const result = await response.json();
			if (
				result.summary &&
				(!summary || result.summary.updatedAt > summary.updatedAt)
			) {
				summary = result.summary;
				summaryStatus = "completed";
				showSummary = true;
				clearPolling();
				if (isUpdate) {
					triggerSummaryUpdateAnimation();
				}
				dispatch("summaryUpdated", { summary });
			}
		}
	} catch (error) {
		console.error("Failed to poll summary status:", error);
	}
}

function clearPolling() {
	if (pollingInterval) {
		clearInterval(pollingInterval);
		pollingInterval = null;
	}
}

// 要約更新時のアニメーション
function triggerSummaryUpdateAnimation() {
	summaryJustUpdated = true;
	setTimeout(() => {
		summaryJustUpdated = false;
	}, 1500);
}

// 要約生成
async function generateSummary() {
	if (summaryGenerating || isDisabled) return;

	summaryGenerating = true;
	summaryStatus = "queued";

	try {
		const response = await authenticatedFetch(generateUrl, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(generatePayload),
		});

		if (response.ok) {
			const result = await response.json();
			if (result.summary) {
				summary = result.summary;
				summaryStatus = "completed";
				showSummary = true;
			} else {
				summaryStatus = "queued";
				startPolling(true);
			}
		} else {
			const errorData = await response.json().catch(() => ({}));
			handleError(errorData, response.status);
		}
	} catch (error) {
		console.error("Failed to generate summary:", error);
		dispatch("error", {
			message:
				type === "daily"
					? $_("diary.summaryGenerationFailed")
					: $_("monthly.summary.error"),
		});
		summaryStatus = "none";
	} finally {
		summaryGenerating = false;
	}
}

function handleError(errorData: any, status: number) {
	let errorMessage = "";

	if (status === 404) {
		errorMessage =
			type === "daily"
				? $_("diary.summaryGenerationFailed")
				: $_("monthly.summary.noEntries");
	} else if (status === 400 && errorData.message?.includes("API key")) {
		errorMessage =
			type === "daily"
				? $_("diary.summaryGenerationFailed")
				: $_("monthly.summary.noApiKey");
	} else if (status === 400 && errorData.message?.includes("current month")) {
		errorMessage = $_("monthly.summary.currentMonthError");
	} else {
		errorMessage =
			type === "daily"
				? $_("diary.summaryGenerationFailed")
				: $_("monthly.summary.error");
	}

	dispatch("error", { message: errorMessage });
	summaryStatus = "none";
}

function startPolling(isUpdate = false) {
	clearPolling();
	pollingInterval = setInterval(() => pollSummaryStatus(isUpdate), 3000);
}

function toggleSummary() {
	showSummary = !showSummary;
}

onDestroy(() => {
	clearPolling();
});

// プロパティ変更時の処理
$: if (summary) {
	summaryStatus = "completed";
}
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
	<div class="flex items-center justify-between mb-4">
		<h2 class="text-xl font-semibold text-gray-900 dark:text-white">
			{type === "daily" ? $_("diary.summary.title") : $_("monthly.summary.title")}
		</h2>

		{#if summary && showSummary}
			<button
				type="button"
				on:click={toggleSummary}
				class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-200 text-sm font-medium"
			>
				{$_("diary.summary.hide")}
			</button>
		{:else if summary}
			<button
				type="button"
				on:click={toggleSummary}
				class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-200 text-sm font-medium"
			>
				{type === "daily" ? $_("diary.summary.view") : $_("monthly.summary.view")}
			</button>
		{/if}
	</div>

	{#if isDisabled}
		<div class="text-center py-4">
			<p class="text-gray-500 dark:text-gray-400 text-sm">
				{disabledMessage}
			</p>
		</div>
	{:else if !hasLLMKey}
		<div class="text-center py-4">
			<p class="text-red-600 dark:text-red-400 text-sm">
				{$_("monthly.summary.noApiKey")}
			</p>
		</div>
	{:else}
		<div class="space-y-4">
			{#if summaryStatus === 'queued'}
				<div class="flex items-center text-blue-600 dark:text-blue-400">
					<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 dark:border-blue-400 mr-2"></div>
					<span>{$_("diary.summary.statusQueued")}</span>
				</div>
			{:else if summaryStatus === 'processing'}
				<div class="flex items-center text-blue-600 dark:text-blue-400">
					<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 dark:border-blue-400 mr-2"></div>
					<span>{$_("diary.summary.statusProcessing")}</span>
				</div>
			{:else if summary && showSummary}
				<div class="prose dark:prose-invert max-w-none">
					<p class="text-gray-700 dark:text-gray-300 whitespace-pre-wrap leading-relaxed transition-all duration-300 px-2 py-1 rounded"
					   class:summary-highlight={summaryJustUpdated}>
						{summary.summary}
					</p>
				</div>
			{/if}

			<div class="flex items-center gap-4">
				{#if summary}
					<button
						type="button"
						on:click={generateSummary}
						disabled={summaryGenerating}
						class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						{#if summaryGenerating}
							{type === "daily" ? $_("diary.regeneratingSummary") : $_("monthly.summary.generating")}
						{:else}
							{type === "daily" ? $_("diary.regenerateSummary") : $_("monthly.summary.regenerate")}
						{/if}
					</button>
				{:else}
					<button
						type="button"
						on:click={generateSummary}
						disabled={summaryGenerating}
						class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
					>
						{#if summaryGenerating}
							{type === "daily" ? $_("diary.generatingSummary") : $_("monthly.summary.generating")}
						{:else}
							{type === "daily" ? $_("diary.generateSummary") : $_("monthly.summary.generate")}
						{/if}
					</button>
				{/if}
			</div>

			{#if summary}
				<div class="text-xs text-gray-500 dark:text-gray-400 space-y-1">
					<div>{$_("common.createdAt")}: {new Date(summary.createdAt).toLocaleString($locale)}</div>
					<div>{$_("common.updatedAt")}: {new Date(summary.updatedAt).toLocaleString($locale)}</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
.summary-highlight {
	box-shadow: inset 0 0 0 2px #3b82f6;
	animation: highlight-pulse 1.5s ease-in-out;
}

@keyframes highlight-pulse {
	0% {
		box-shadow: inset 0 0 0 2px #3b82f6;
	}
	50% {
		box-shadow: inset 0 0 0 4px #60a5fa;
	}
	100% {
		box-shadow: inset 0 0 0 2px #3b82f6;
	}
}
</style>