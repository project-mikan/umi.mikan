<script lang="ts">
import { _, locale } from "svelte-i18n";
import { browser } from "$app/environment";
import { onMount } from "svelte";
import { authenticatedFetch } from "$lib/auth-client";
import "$lib/i18n";

interface LatestTrendData {
	analysis: string;
	periodStart: string;
	periodEnd: string;
	generatedAt: string;
}

export let userName: string | null = null;

let trendData: LatestTrendData | null = null;
let isLoading = true;
let errorMessage = "";

// トレンド分析データを取得
async function fetchLatestTrend(retryCount = 0) {
	if (!browser) return;

	isLoading = true;
	errorMessage = "";

	try {
		const response = await authenticatedFetch("/api/diary/latest-trend");
		if (response.ok) {
			const result = await response.json();
			if (result.analysis) {
				trendData = {
					analysis: result.analysis,
					periodStart: result.periodStart,
					periodEnd: result.periodEnd,
					generatedAt: result.generatedAt,
				};
			} else {
				// データが空の場合はnullにする
				trendData = null;
			}
		} else if (response.status === 404) {
			// 404の場合はデータが存在しない
			trendData = null;
		} else if (response.status >= 500 && retryCount < 2) {
			// サーバーエラーの場合は最大2回リトライ
			console.warn(`Server error, retrying... (${retryCount + 1}/2)`);
			setTimeout(
				() => fetchLatestTrend(retryCount + 1),
				1000 * (retryCount + 1),
			);
			return;
		} else {
			errorMessage = $_("latestTrend.error");
		}
	} catch (error) {
		console.error("Failed to fetch latest trend:", error);
		if (retryCount < 2) {
			// ネットワークエラーの場合も最大2回リトライ
			console.warn(`Network error, retrying... (${retryCount + 1}/2)`);
			setTimeout(
				() => fetchLatestTrend(retryCount + 1),
				1000 * (retryCount + 1),
			);
			return;
		}
		errorMessage = $_("latestTrend.error");
	} finally {
		isLoading = false;
	}
}

// 期間の日本語表示を生成
function formatPeriod(start: string, end: string): string {
	if (!start || !end) return "";

	const startDate = new Date(start);
	const endDate = new Date(end);

	if ($locale === "ja") {
		const startStr = `${startDate.getFullYear()}年${startDate.getMonth() + 1}月${startDate.getDate()}日`;
		const endStr = `${endDate.getFullYear()}年${endDate.getMonth() + 1}月${endDate.getDate()}日`;
		return `${startStr} 〜 ${endStr}`;
	} else {
		return `${startDate.toLocaleDateString($locale || "en")} - ${endDate.toLocaleDateString($locale || "en")}`;
	}
}

// 分析テキストをHTML形式に変換（見出しマーカーをスタイリング）
function formatAnalysis(text: string): string {
	if (!text) return "";

	return text
		.replace(
			/^## (.+)$/gm,
			'<h2 class="text-lg font-bold text-gray-900 dark:text-white mt-4 mb-2">$1</h2>',
		)
		.replace(
			/^### (.+)$/gm,
			'<h3 class="text-base font-semibold text-gray-800 dark:text-gray-200 mt-3 mb-1">$1</h3>',
		)
		.replace(/\n/g, "<br>");
}

onMount(() => {
	fetchLatestTrend();
});
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
	<div class="flex items-center justify-between mb-4">
		<div>
			<h2 class="text-xl font-semibold text-gray-900 dark:text-white">
				{userName ? $_("latestTrend.titleWithName", { values: { name: userName } }) : $_("latestTrend.title")}
			</h2>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
				{$_("latestTrend.description")}
			</p>
		</div>
	</div>

	{#if isLoading}
		<div class="flex items-center justify-center py-8">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 dark:border-blue-400"></div>
			<span class="ml-3 text-gray-600 dark:text-gray-300">{$_("latestTrend.loading")}</span>
		</div>
	{:else if errorMessage}
		<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
			<p class="text-red-800 dark:text-red-200">{errorMessage}</p>
		</div>
	{:else if trendData}
		<div class="space-y-4">
			<!-- 分析期間 -->
			<div class="text-sm text-gray-600 dark:text-gray-400">
				<span class="font-medium">{$_("latestTrend.period")}:</span>
				{formatPeriod(trendData.periodStart, trendData.periodEnd)}
			</div>

			<!-- 分析内容 -->
			<div class="prose dark:prose-invert max-w-none">
				<div class="text-gray-700 dark:text-gray-300 leading-relaxed bg-gray-50 dark:bg-gray-700/50 rounded-md p-4 auto-phrase-target">
					{@html formatAnalysis(trendData.analysis)}
				</div>
			</div>

			<!-- 生成日時 -->
			<div class="text-xs text-gray-500 dark:text-gray-400">
				{$_("latestTrend.generatedAt")}: {new Date(trendData.generatedAt).toLocaleString($locale || "en")}
			</div>
		</div>
	{:else}
		<div class="text-center py-8">
			<p class="text-gray-500 dark:text-gray-400">
				{$_("latestTrend.noData")}
			</p>
			<p class="text-sm text-gray-400 dark:text-gray-500 mt-2">
				{$_("latestTrend.notEnoughData")}
			</p>
		</div>
	{/if}
</div>
