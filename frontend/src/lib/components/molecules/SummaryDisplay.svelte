<script lang="ts">
import { _, locale } from "svelte-i18n";
import { browser } from "$app/environment";
import { onMount, onDestroy, createEventDispatcher } from "svelte";
import { authenticatedFetch } from "$lib/auth-client";
import "$lib/i18n";

interface Summary {
	id: string;
	summary: string;
	createdAt: number;
	updatedAt: number;
}

export let type: "daily" | "monthly"; // 日次または月次
export let summary: Summary | null = null;
export let fetchUrl: string; // 要約取得用URL
export let generateUrl: string; // 要約生成用URL
export let generatePayload: Record<string, unknown> = {}; // 生成時に送信するペイロード
export let isDisabled = false; // 生成ボタンを無効にするかどうか
export let disabledMessage = ""; // 無効時のメッセージ
export let showSummary = true;
export let hasLLMKey = true;

const dispatch = createEventDispatcher();

let summaryStatus: "none" | "queued" | "processing" | "completed" = summary
	? "completed"
	: "none";
let summaryGenerating = false;
let isRegenerating = false; // 再生成かどうかのフラグ
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
				summaryGenerating = false; // ポーリング完了時にローディング終了
				clearPolling();
				if (isUpdate || isRegenerating) {
					triggerSummaryUpdateAnimation();
				}
				isRegenerating = false;
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

	isRegenerating = summary !== null; // 既に要約がある場合は再生成
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
			if (result.summary?.summary) {
				const summaryText = result.summary.summary;

				// ステータスメッセージのチェック（要約として不完全な場合）
				if (
					summaryText.includes("queued") ||
					summaryText.includes("Please check back later") ||
					summaryText.includes("generation has been queued")
				) {
					summaryStatus = "queued";
					startPolling(true);
					// ポーリング中はローディング状態を維持
				} else if (
					summaryText.includes("processing") ||
					summaryText.includes("Updating") ||
					summaryText.includes("generating")
				) {
					summaryStatus = "processing";
					startPolling(true);
					// ポーリング中はローディング状態を維持
				} else {
					// 正常な要約が完成
					summary = result.summary;
					summaryStatus = "completed";
					showSummary = true;
					summaryGenerating = false;
					if (isRegenerating) {
						triggerSummaryUpdateAnimation();
					}
					isRegenerating = false;
					dispatch("summaryUpdated", { summary });
				}
			} else {
				summaryStatus = "queued";
				startPolling(true);
				// ポーリング中はローディング状態を維持
			}
		} else {
			const errorData = await response.json().catch(() => ({}));
			handleError(errorData, response.status);
			summaryGenerating = false;
			isRegenerating = false;
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
		summaryGenerating = false;
		isRegenerating = false;
	}
}

function handleError(errorData: Record<string, unknown>, status: number) {
	let errorMessage = "";

	if (status === 404) {
		errorMessage =
			type === "daily"
				? $_("diary.summaryGenerationFailed")
				: $_("monthly.summary.noEntries");
	} else if (
		status === 400 &&
		typeof errorData.message === "string" &&
		errorData.message.includes("API key")
	) {
		errorMessage =
			type === "daily"
				? $_("diary.summaryGenerationFailed")
				: $_("monthly.summary.noApiKey");
	} else if (
		status === 400 &&
		typeof errorData.message === "string" &&
		errorData.message.includes("current month")
	) {
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

// 既存要約を取得
async function fetchExistingSummary() {
	if (!browser) return;

	try {
		const response = await authenticatedFetch(fetchUrl);
		if (response.ok) {
			const result = await response.json();
			if (result.summary) {
				summary = result.summary;
				summaryStatus = "completed";
				showSummary = true;
				dispatch("summaryUpdated", { summary });
			}
		} else if (response.status !== 404) {
			console.error("Failed to fetch summary:", response.status);
		}
	} catch (error) {
		console.error("Failed to fetch summary:", error);
	}
}

// ポーリング開始
function startPolling(isUpdate = false) {
	clearPolling();
	pollingInterval = setInterval(() => pollSummaryStatus(isUpdate), 3000);
}

function toggleSummary() {
	showSummary = !showSummary;
}

onMount(() => {
	// コンポーネント初期化時に既存要約を取得（無効化されていない場合のみ）
	if (!isDisabled) {
		fetchExistingSummary();
	}
});

onDestroy(() => {
	clearPolling();
});

// プロパティ変更時の処理
$: if (summary) {
	summaryStatus = "completed";
}

// fetchUrl または generatePayload が変更された時の処理（月変更時）
let previousFetchUrl = "";
let previousGeneratePayload = "";

$: {
	const currentPayload = JSON.stringify(generatePayload);
	if (
		browser &&
		previousFetchUrl &&
		(fetchUrl !== previousFetchUrl ||
			currentPayload !== previousGeneratePayload)
	) {
		// 状態をリセット
		summary = null;
		summaryStatus = "none";
		summaryGenerating = false;
		// showSummaryは無効化状態でも表示が必要なのでリセットしない
		// showSummary = false;
		clearPolling();

		// 新しいURL/パラメータで要約を取得（無効化されていない場合のみ）
		if (!isDisabled) {
			fetchExistingSummary();
		}
	}

	// 前回の値を更新（初回も含む）
	previousFetchUrl = fetchUrl;
	previousGeneratePayload = currentPayload;
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
				on:click|preventDefault|stopPropagation={toggleSummary}
				class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-200 text-sm font-medium"
			>
				{$_("diary.summary.hide")}
			</button>
		{:else if summary}
			<button
				type="button"
				on:click|preventDefault|stopPropagation={toggleSummary}
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
			{#if summaryGenerating}
				<div class="flex items-center text-blue-600 dark:text-blue-400">
					<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 dark:border-blue-400 mr-2"></div>
					<span>
						{#if isRegenerating}
							{type === "daily" ? $_("diary.regeneratingSummary") : $_("monthly.summary.regenerating")}
						{:else}
							{type === "daily" ? $_("diary.generatingSummary") : $_("monthly.summary.generating")}
						{/if}
					</span>
				</div>
			{:else if summaryStatus === 'queued'}
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
						on:click|preventDefault|stopPropagation={generateSummary}
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
						on:click|preventDefault|stopPropagation={generateSummary}
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
					<div>{$_("common.createdAt")}: {new Date(summary.createdAt).toLocaleString($locale || "en")}</div>
					<div>{$_("common.updatedAt")}: {new Date(summary.updatedAt).toLocaleString($locale || "en")}</div>
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