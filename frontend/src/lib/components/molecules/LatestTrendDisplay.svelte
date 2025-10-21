<script lang="ts">
import { _, locale } from "svelte-i18n";
import { browser } from "$app/environment";
import { onMount } from "svelte";
import { authenticatedFetch } from "$lib/auth-client";
import { summaryVisibility } from "$lib/summary-visibility-store";
import "$lib/i18n";

interface LatestTrendData {
	overallSummary: string;
	healthMood: string;
	activities: string;
	concerns: string;
	periodStart: string;
	periodEnd: string;
	generatedAt: string;
}

export let userName: string | null = null;

let trendData: LatestTrendData | null = null;
let isLoading = true;
let errorMessage = "";

// ストアから表示状態を取得
$: showTrend = $summaryVisibility.latestTrend;

function toggleTrend() {
	summaryVisibility.toggleLatestTrend();
}

// トレンド分析データを取得
async function fetchLatestTrend(retryCount = 0) {
	if (!browser) return;

	isLoading = true;
	errorMessage = "";

	try {
		const response = await authenticatedFetch("/api/diary/latest-trend");
		if (response.ok) {
			const result = await response.json();

			// データのバリデーション
			if (
				result.overallSummary &&
				typeof result.overallSummary === "string" &&
				result.healthMood &&
				typeof result.healthMood === "string" &&
				result.activities &&
				typeof result.activities === "string" &&
				result.concerns &&
				typeof result.concerns === "string" &&
				result.periodStart &&
				result.periodEnd &&
				result.generatedAt
			) {
				// 日付の妥当性チェック
				const startDate = new Date(result.periodStart);
				const endDate = new Date(result.periodEnd);
				const generatedDate = new Date(result.generatedAt);

				if (
					!Number.isNaN(startDate.getTime()) &&
					!Number.isNaN(endDate.getTime()) &&
					!Number.isNaN(generatedDate.getTime())
				) {
					trendData = {
						overallSummary: result.overallSummary,
						healthMood: result.healthMood,
						activities: result.activities,
						concerns: result.concerns,
						periodStart: result.periodStart,
						periodEnd: result.periodEnd,
						generatedAt: result.generatedAt,
					};
				} else {
					// 日付が不正な場合
					console.warn("Invalid date format in latest trend data");
					trendData = null;
				}
			} else {
				// データが空または不正な場合はnullにする
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

onMount(() => {
	// summaryVisibility.init()は+layout.svelteで既に呼ばれているため不要
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
		{#if trendData && showTrend}
			<button
				type="button"
				on:click|preventDefault|stopPropagation={toggleTrend}
				class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-200 text-sm font-medium"
			>
				{$_("latestTrend.hide")}
			</button>
		{:else if trendData}
			<button
				type="button"
				on:click|preventDefault|stopPropagation={toggleTrend}
				class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-200 text-sm font-medium"
			>
				{$_("latestTrend.view")}
			</button>
		{/if}
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
	{:else if trendData && showTrend}
		<div class="space-y-4">
			<!-- 分析期間 -->
			<div class="text-sm text-gray-600 dark:text-gray-400">
				<span class="font-medium">{$_("latestTrend.period")}:</span>
				{formatPeriod(trendData.periodStart, trendData.periodEnd)}
			</div>

			<!-- 全体的な様子 -->
			<div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-4">
				<h2 class="text-lg font-bold text-gray-900 dark:text-white mb-2">
					{$_("latestTrend.overallSummary")}
				</h2>
				<p class="text-gray-700 dark:text-gray-300 leading-relaxed auto-phrase-target">
					{trendData.overallSummary}
				</p>
			</div>

			<!-- 体調・気分 -->
			<div class="bg-green-50 dark:bg-green-900/20 rounded-lg p-4">
				<h3 class="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">
					{$_("latestTrend.healthMood")}
				</h3>
				<p class="text-gray-700 dark:text-gray-300 leading-relaxed auto-phrase-target">
					{trendData.healthMood}
				</p>
			</div>

			<!-- 活動・行動 -->
			<div class="bg-purple-50 dark:bg-purple-900/20 rounded-lg p-4">
				<h3 class="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">
					{$_("latestTrend.activities")}
				</h3>
				<p class="text-gray-700 dark:text-gray-300 leading-relaxed auto-phrase-target">
					{trendData.activities}
				</p>
			</div>

			<!-- 気になること -->
			<div class="bg-orange-50 dark:bg-orange-900/20 rounded-lg p-4">
				<h3 class="text-base font-semibold text-gray-800 dark:text-gray-200 mb-2">
					{$_("latestTrend.concerns")}
				</h3>
				<p class="text-gray-700 dark:text-gray-300 leading-relaxed auto-phrase-target">
					{trendData.concerns}
				</p>
			</div>

			<!-- 生成日時 -->
			<div class="text-xs text-gray-500 dark:text-gray-400">
				{$_("latestTrend.generatedAt")}: {new Date(trendData.generatedAt).toLocaleString($locale || "en")}
			</div>
		</div>
	{:else if !trendData}
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
