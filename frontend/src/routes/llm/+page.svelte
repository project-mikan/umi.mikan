<script lang="ts">
  import { _, locale } from "svelte-i18n";
  import "$lib/i18n";
  import Head from "$lib/components/atoms/Head.svelte";
  import Card from "$lib/components/atoms/Card.svelte";
  import PubSubMetricsChart from "$lib/components/molecules/PubSubMetricsChart.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ metrics } = data);

  // 処理中タスクの表示用フォーマット
  function formatProcessingTask(task: {
    taskType: string;
    date: string;
    startedAt: number;
  }) {
    let type = "";
    if (task.taskType === "daily_summary") {
      type = $_("llm.metrics.dailySummary");
    } else if (task.taskType === "monthly_summary") {
      type = $_("llm.metrics.monthlySummary");
    } else if (task.taskType === "latest_trend") {
      type = $_("llm.metrics.latestTrend");
    }
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

  // インデックス生成フローのステップ（AI処理に注目）
  $: indexingSteps = [
    {
      icon: "📄",
      label: $_("llm.ragDiagram.step.diaryText"),
      subLabel: null as string | null,
      style:
        "bg-gray-50 dark:bg-gray-700/50 border-gray-200 dark:border-gray-600 text-gray-800 dark:text-gray-200",
    },
    {
      icon: "🧹",
      label: $_("llm.ragDiagram.step.toPlainText"),
      subLabel: $_("llm.ragDiagram.subLabel.removeMarkdown") as string | null,
      style:
        "bg-slate-50 dark:bg-slate-800/50 border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-200",
    },
    {
      icon: "✂️",
      label: $_("llm.ragDiagram.step.splitByTopic"),
      subLabel: $_("llm.ragDiagram.subLabel.llmSplitWithSummary") as
        | string
        | null,
      style:
        "bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800 text-orange-800 dark:text-orange-200",
    },
    {
      icon: "📅",
      label: $_("llm.ragDiagram.step.enrichWithDate"),
      subLabel: $_("llm.ragDiagram.subLabel.datePrefix") as string | null,
      style:
        "bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-200",
    },
    {
      icon: "🔢",
      label: $_("llm.ragDiagram.step.vectorizeDocument"),
      subLabel: $_("llm.ragDiagram.subLabel.retrievalDocument") as
        | string
        | null,
      style:
        "bg-purple-50 dark:bg-purple-900/20 border-purple-200 dark:border-purple-800 text-purple-800 dark:text-purple-200",
    },
    {
      icon: "🗄️",
      label: $_("llm.ragDiagram.step.saveToVectorDb"),
      subLabel: $_("llm.ragDiagram.subLabel.hnswStorage") as string | null,
      style:
        "bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800 text-blue-800 dark:text-blue-200",
    },
  ];

  // 検索フローのステップ（AI処理に注目）
  $: searchSteps = [
    {
      icon: "🔍",
      label: $_("llm.ragDiagram.step.naturalLanguageQuery"),
      subLabel: $_("llm.ragDiagram.subLabel.queryExample") as string | null,
      style:
        "bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-800 dark:text-green-200",
    },
    {
      icon: "🔢",
      label: $_("llm.ragDiagram.step.vectorizeQuery"),
      subLabel: $_("llm.ragDiagram.subLabel.retrievalQuery") as string | null,
      style:
        "bg-purple-50 dark:bg-purple-900/20 border-purple-200 dark:border-purple-800 text-purple-800 dark:text-purple-200",
    },
    {
      icon: "⚡",
      label: $_("llm.ragDiagram.step.searchSimilarChunks"),
      subLabel: $_("llm.ragDiagram.subLabel.cosineSimilarity") as string | null,
      style:
        "bg-indigo-50 dark:bg-indigo-900/20 border-indigo-200 dark:border-indigo-800 text-indigo-800 dark:text-indigo-200",
    },
    {
      icon: "📖",
      label: $_("llm.ragDiagram.step.keywordFallback"),
      subLabel: $_("llm.ragDiagram.subLabel.properNounFallback") as
        | string
        | null,
      style:
        "bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-200",
    },
    {
      icon: "✅",
      label: $_("llm.ragDiagram.step.returnWithSummary"),
      subLabel: $_("llm.ragDiagram.subLabel.withScore") as string | null,
      style:
        "bg-teal-50 dark:bg-teal-900/20 border-teal-200 dark:border-teal-800 text-teal-800 dark:text-teal-200",
    },
  ];

  // RAGカード用データ
  $: ragCards = [
    {
      title: $_("llm.metrics.totalEmbeddings"),
      value: metrics.summary.totalEmbeddings,
      color: "text-purple-600 dark:text-purple-400",
    },
    {
      title: $_("llm.metrics.pendingEmbeddings"),
      value: metrics.summary.pendingEmbeddings,
      color: "text-pink-600 dark:text-pink-400",
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
		<p class="text-gray-600 dark:text-gray-400 auto-phrase-target">
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
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
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
			<div class="flex items-center justify-between">
				<span class="text-gray-700 dark:text-gray-300">
					{$_("llm.metrics.latestTrendAutoGeneration")}
				</span>
				<span class="px-2 py-1 rounded-full text-xs font-medium {
					metrics.summary.autoLatestTrendEnabled
						? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
						: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
				}">
					{metrics.summary.autoLatestTrendEnabled ? $_("common.enabled") : $_("common.disabled")}
				</span>
			</div>
		</div>
	</Card>

	<!-- RAG（自然言語検索）処理状況 -->
	<Card>
		<div class="flex items-center justify-between mb-4">
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white">
				{$_("llm.metrics.ragTitle")}
			</h3>
			<span class="px-2 py-1 rounded-full text-xs font-medium {
				metrics.summary.semanticSearchEnabled
					? 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200'
					: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
			}">
				{metrics.summary.semanticSearchEnabled ? $_("common.enabled") : $_("common.disabled")}
			</span>
		</div>
		{#if metrics.summary.semanticSearchEnabled}
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
				{#each ragCards as card}
					<div class="text-center p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
						<div class="text-2xl font-bold {card.color}">
							{card.value}
						</div>
						<div class="text-sm text-gray-600 dark:text-gray-400 mt-1">
							{card.title}
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<p class="text-sm text-gray-500 dark:text-gray-400 auto-phrase-target">
				{$_("llm.metrics.ragDisabledNote")}
			</p>
		{/if}
	</Card>

	<!-- トレンド生成状況 -->
	{#if metrics.summary.latestTrendGeneratedAt}
		<Card>
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
				{$_("llm.metrics.latestTrendStatus")}
			</h3>
			<div class="flex items-center justify-between">
				<span class="text-gray-700 dark:text-gray-300">
					{$_("llm.metrics.lastGeneratedAt")}
				</span>
				<span class="text-sm text-gray-600 dark:text-gray-400">
					{new Date(metrics.summary.latestTrendGeneratedAt).toLocaleString($locale || "ja")}
				</span>
			</div>
		</Card>
	{/if}

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
			<p class="text-gray-600 dark:text-gray-400 auto-phrase-target">
				{$_("llm.metrics.aboutDescription")}
			</p>
			<ul class="text-gray-600 dark:text-gray-400 mt-2 auto-phrase-target">
				<li>{$_("llm.metrics.dailySummaryExplanation")}</li>
				<li>{$_("llm.metrics.monthlySummaryExplanation")}</li>
				<li>{$_("llm.metrics.processingExplanation")}</li>
			</ul>
		</div>

		<!-- RAGの仕組み図 -->
		<div class="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
			<h4 class="text-base font-semibold text-gray-900 dark:text-white mb-5">
				{$_("llm.ragDiagram.title")}
			</h4>

			<!-- インデックス生成フロー -->
			<div class="mb-6">
				<div class="flex items-center gap-2 mb-3">
					<span class="px-2 py-1 bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 text-xs font-medium rounded-full">
						{$_("llm.ragDiagram.asyncBadge")}
					</span>
					<h5 class="text-sm font-medium text-gray-700 dark:text-gray-300">
						{$_("llm.ragDiagram.indexingTitle")}
					</h5>
				</div>
				<div class="overflow-x-auto pb-2">
					<div class="flex items-center gap-1 min-w-max">
						{#each indexingSteps as step, i}
							<div class="flex flex-col items-center justify-start w-28 min-h-[88px] p-2 rounded-lg border text-center overflow-hidden {step.style}">
								<span class="text-xl mb-1 shrink-0">{step.icon}</span>
								<span class="text-xs font-medium leading-tight break-words w-full">{step.label}</span>
								{#if step.subLabel}
									<span class="text-xs opacity-70 mt-0.5 leading-tight break-words w-full" style="white-space: pre-line">{step.subLabel}</span>
								{/if}
							</div>
							{#if i < indexingSteps.length - 1}
								<svg class="w-4 h-4 text-gray-400 dark:text-gray-500 shrink-0" fill="currentColor" viewBox="0 0 20 20">
									<path fill-rule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clip-rule="evenodd" />
								</svg>
							{/if}
						{/each}
					</div>
				</div>
			</div>

			<!-- 検索フロー -->
			<div>
				<div class="flex items-center gap-2 mb-3">
					<span class="px-2 py-1 bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200 text-xs font-medium rounded-full">
						{$_("llm.ragDiagram.syncBadge")}
					</span>
					<h5 class="text-sm font-medium text-gray-700 dark:text-gray-300">
						{$_("llm.ragDiagram.searchTitle")}
					</h5>
				</div>
				<div class="overflow-x-auto pb-2">
					<div class="flex items-center gap-1 min-w-max">
						{#each searchSteps as step, i}
							<div class="flex flex-col items-center justify-start w-28 min-h-[88px] p-2 rounded-lg border text-center overflow-hidden {step.style}">
								<span class="text-xl mb-1 shrink-0">{step.icon}</span>
								<span class="text-xs font-medium leading-tight break-words w-full">{step.label}</span>
								{#if step.subLabel}
									<span class="text-xs opacity-70 mt-0.5 leading-tight break-words w-full" style="white-space: pre-line">{step.subLabel}</span>
								{/if}
							</div>
							{#if i < searchSteps.length - 1}
								<svg class="w-4 h-4 text-gray-400 dark:text-gray-500 shrink-0" fill="currentColor" viewBox="0 0 20 20">
									<path fill-rule="evenodd" d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z" clip-rule="evenodd" />
								</svg>
							{/if}
						{/each}
					</div>
				</div>
			</div>
		</div>
	</Card>
</div>