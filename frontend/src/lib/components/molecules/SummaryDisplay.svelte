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
export let isGenerating = false; // 親コンポーネントから生成状況を受け取る
export let isSummaryOutdated = false; // 要約が古いかどうか

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
			if (result.summary) {
				const summaryText = result.summary.summary;

				// ステータスメッセージのチェック（まだ生成中かどうか判定）
				if (
					summaryText.includes("queued") ||
					summaryText.includes("Please check back later") ||
					summaryText.includes("generation has been queued")
				) {
					// まだキューイング中
					summaryStatus = "queued";
					// ポーリング継続
				} else if (
					summaryText.includes("processing") ||
					summaryText.includes("Updating") ||
					summaryText.includes("generating")
				) {
					// まだ処理中
					summaryStatus = "processing";
					// ポーリング継続
				} else {
					// 正常な要約が完成した
					const newSummary = result.summary;
					const oldSummary = summary;

					// 要約が実際に更新されたかどうかを確認
					// ポーリング時は初回取得(!oldSummary)をアニメーション対象外とする
					const actuallyUpdated =
						oldSummary &&
						(oldSummary.updatedAt !== newSummary.updatedAt ||
						oldSummary.summary !== newSummary.summary);

					// 実際に更新された場合、または初回取得の場合のみsummaryを更新
					if (actuallyUpdated || !summary) {
						summary = newSummary;
						summaryStatus = "completed";
						showSummary = true;
						summaryGenerating = false; // ポーリング完了時にローディング終了
						clearPolling();

						// 実際に更新された場合で、かつ再生成中の場合のみアニメーションを発火
						if (actuallyUpdated && isRegenerating) {
							triggerSummaryUpdateAnimation();
						}
						isRegenerating = false;

						// 実際に更新された場合のみイベントを発火
						if (actuallyUpdated) {
							dispatch("summaryUpdated", { summary: newSummary });
						}
						dispatch("generationCompleted");
					} else {
						// 同じ内容の場合はポーリング継続
						console.log("Polling: Same summary content received, continuing polling");
					}
				}
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
	// デバッグ用ログ（開発環境でのみ）
	if (typeof window !== "undefined" && window.location.hostname === "localhost") {
		console.log("🎯 Animation triggered!", {
			isRegenerating,
			summaryGenerating,
			summaryStatus,
			timestamp: new Date().toISOString(),
			stackTrace: new Error().stack?.split('\n').slice(1, 4).join('\n')
		});
	}

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

	// 生成開始をイベントで通知
	dispatch("generationStarted");

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
				} else if (
					summaryText.includes("processing") ||
					summaryText.includes("Updating") ||
					summaryText.includes("generating")
				) {
					summaryStatus = "processing";
					startPolling(true);
				} else {
					// 正常な要約が完成
					const newSummary = result.summary;
					const oldSummary = summary;

					// 要約が実際に更新されたかどうかを確認
					// 要約生成時も初回取得(!oldSummary)を更新対象外とする
					const actuallyUpdated =
						oldSummary &&
						(oldSummary.updatedAt !== newSummary.updatedAt ||
						oldSummary.summary !== newSummary.summary);

					// 実際に更新された場合、または初回取得の場合のみsummaryを更新
					if (actuallyUpdated || !summary) {
						summary = newSummary;
						summaryStatus = "completed";
						showSummary = true;
						summaryGenerating = false;
						isRegenerating = false;

						// 実際に更新された場合のみイベントを発火
						if (actuallyUpdated) {
							dispatch("summaryUpdated", { summary: newSummary });
						}
						dispatch("generationCompleted");
					} else {
						// 同じ内容の場合は状態を変更せずにポーリング継続
						// 実際の新しい要約が生成されるまで待機
						console.log("Same summary content received, continuing to wait for actual update");
					}
				}
			} else {
				summaryStatus = "queued";
				startPolling(true);
			}
		} else {
			const errorData = await response.json().catch(() => ({}));
			handleError(errorData, response.status);
			summaryGenerating = false;
			isRegenerating = false;
			dispatch("generationCompleted");
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
		dispatch("generationCompleted");
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
				const summaryText = result.summary.summary;

				// ステータスメッセージのチェック（生成中かどうか判定）
				if (
					summaryText.includes("queued") ||
					summaryText.includes("Please check back later") ||
					summaryText.includes("generation has been queued")
				) {
					summaryStatus = "queued";
					summaryGenerating = true;
					startPolling(true);
					dispatch("generationStarted");
				} else if (
					summaryText.includes("processing") ||
					summaryText.includes("Updating") ||
					summaryText.includes("generating")
				) {
					summaryStatus = "processing";
					summaryGenerating = true;
					startPolling(true);
					dispatch("generationStarted");
				} else {
					// 正常な要約が存在
					summary = result.summary;
					summaryStatus = "completed";
					showSummary = true;
					summaryGenerating = false;
					dispatch("summaryUpdated", { summary });
				}
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
$: {
	if (isGenerating) {
		summaryGenerating = true;
		summaryStatus = "processing";
	} else if (summary) {
		summaryGenerating = false;
		summaryStatus = "completed";
	} else {
		summaryGenerating = false;
		summaryStatus = "none";
	}
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
					{#if isSummaryOutdated}
						<div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md p-3 mb-4">
							<div class="flex">
								<div class="flex-shrink-0">
									<svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
									</svg>
								</div>
								<div class="ml-3">
									<p class="text-sm text-yellow-800 dark:text-yellow-200">
										{type === "daily" ? $_("diary.summary.outdated") : $_("monthly.summary.outdated")}
									</p>
								</div>
							</div>
						</div>
					{/if}
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