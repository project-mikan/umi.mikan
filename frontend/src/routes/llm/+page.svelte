<script lang="ts">
import { _, locale } from "svelte-i18n";
import "$lib/i18n";
import Head from "$lib/components/atoms/Head.svelte";
import Card from "$lib/components/atoms/Card.svelte";
import PubSubMetricsChart from "$lib/components/molecules/PubSubMetricsChart.svelte";
import { autoPhraseEnabled } from "$lib/auto-phrase-store";
import type { PageData } from "./$types";

export let data: PageData;

$: ({ metrics } = data);

// 処理中タスクの表示用フォーマット
function formatProcessingTask(task: {
	taskType: string;
	date: string;
	startedAt: number;
}) {
	const type =
		task.taskType === "daily_summary"
			? $_("llm.metrics.dailySummary")
			: $_("llm.metrics.monthlySummary");
	const startedAt = new Date(task.startedAt * 1000).toLocaleTimeString();
	return `${type} (${task.date}) - ${$_("llm.metrics.startedAt")} ${startedAt}`;
}

// 統計情報のカード用データ
$: summaryCards = [
	{
		title: $_("llm.metrics.totalDailySummaries"),
		value: metrics.summary.totalDailySummaries,
		color: "text-green-600 dark:text-green-400",
	},
	{
		title: $_("llm.metrics.totalMonthlySummaries"),
		value: metrics.summary.totalMonthlySummaries,
		color: "text-blue-600 dark:text-blue-400",
	},
	{
		title: $_("llm.metrics.pendingDailySummaries"),
		value: metrics.summary.pendingDailySummaries,
		color: "text-yellow-600 dark:text-yellow-400",
	},
	{
		title: $_("llm.metrics.pendingMonthlySummaries"),
		value: metrics.summary.pendingMonthlySummaries,
		color: "text-orange-600 dark:text-orange-400",
	},
];
</script>

<Head title={$_("llm.title")} />

<div class="max-w-7xl mx-auto p-4 space-y-6">
	<!-- ページタイトル -->
	<div class="text-center">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
			{$_("llm.title")}
		</h1>
		<p class="text-gray-600 dark:text-gray-400 {$autoPhraseEnabled ? 'auto-phrase' : ''}">
			{$_("llm.description")}
		</p>
	</div>

	<!-- 統計情報カード -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
		{#each summaryCards as card}
			<Card>
				<div class="text-center">
					<div class="text-2xl font-bold {card.color}">
						{card.value}
					</div>
					<div class="text-sm text-gray-600 dark:text-gray-400 mt-1">
						{card.title}
					</div>
				</div>
			</Card>
		{/each}
	</div>

	<!-- 自動要約設定状況 -->
	<Card>
		<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
			{$_("llm.metrics.autoSummarySettings")}
		</h3>
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div class="flex items-center justify-between">
				<span class="text-gray-700 dark:text-gray-300">
					{$_("llm.metrics.dailyAutoSummary")}
				</span>
				<span class="px-2 py-1 rounded-full text-xs font-medium {
					metrics.summary.autoSummaryDailyEnabled
						? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
						: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
				}">
					{metrics.summary.autoSummaryDailyEnabled ? $_("common.enabled") : $_("common.disabled")}
				</span>
			</div>
			<div class="flex items-center justify-between">
				<span class="text-gray-700 dark:text-gray-300">
					{$_("llm.metrics.monthlyAutoSummary")}
				</span>
				<span class="px-2 py-1 rounded-full text-xs font-medium {
					metrics.summary.autoSummaryMonthlyEnabled
						? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
						: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
				}">
					{metrics.summary.autoSummaryMonthlyEnabled ? $_("common.enabled") : $_("common.disabled")}
				</span>
			</div>
		</div>
	</Card>

	<!-- 処理中タスク -->
	{#if metrics.processingTasks.length > 0}
		<Card>
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
				{$_("llm.metrics.processingTasks")}
			</h3>
			<div class="space-y-2">
				{#each metrics.processingTasks as task}
					<div class="flex items-center justify-between p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
						<span class="text-gray-700 dark:text-gray-300">
							{formatProcessingTask(task)}
						</span>
						<div class="flex items-center">
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-yellow-600"></div>
							<span class="ml-2 text-sm text-yellow-600 dark:text-yellow-400">
								{$_("llm.metrics.processing")}
							</span>
						</div>
					</div>
				{/each}
			</div>
		</Card>
	{/if}

	<!-- 24時間の処理状況グラフ -->
	<div>
		<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
			{$_("llm.metrics.last24Hours")}
		</h3>
		<PubSubMetricsChart hourlyMetrics={metrics.hourlyMetrics} />
	</div>

	<!-- 説明テキスト -->
	<Card>
		<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
			{$_("llm.metrics.aboutTitle")}
		</h3>
		<div class="prose dark:prose-invert max-w-none">
			<p class="text-gray-600 dark:text-gray-400 {$autoPhraseEnabled ? 'auto-phrase' : ''}">
				{$_("llm.metrics.aboutDescription")}
			</p>
			<ul class="text-gray-600 dark:text-gray-400 mt-2 {$autoPhraseEnabled ? 'auto-phrase' : ''}">
				<li>{$_("llm.metrics.dailySummaryExplanation")}</li>
				<li>{$_("llm.metrics.monthlySummaryExplanation")}</li>
				<li>{$_("llm.metrics.processingExplanation")}</li>
			</ul>
		</div>
	</Card>
</div>