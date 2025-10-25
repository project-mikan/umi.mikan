<script lang="ts">
	import { onMount, onDestroy } from "svelte";
	import { browser } from "$app/environment";
	import { _, locale, isLoading } from "svelte-i18n";
	import "$lib/i18n";

	// Chart.js の型定義
	let Chart: typeof import("chart.js").Chart;
	let CategoryScale: typeof import("chart.js").CategoryScale;
	let LinearScale: typeof import("chart.js").LinearScale;
	let PointElement: typeof import("chart.js").PointElement;
	let LineElement: typeof import("chart.js").LineElement;
	let LineController: typeof import("chart.js").LineController;
	let BarElement: typeof import("chart.js").BarElement;
	let BarController: typeof import("chart.js").BarController;
	let Title: typeof import("chart.js").Title;
	let Tooltip: typeof import("chart.js").Tooltip;
	let Legend: typeof import("chart.js").Legend;

	export let hourlyMetrics: Array<{
		timestamp: number;
		dailySummariesProcessed: number;
		monthlySummariesProcessed: number;
		dailySummariesFailed: number;
		monthlySummariesFailed: number;
	}>;

	let chartCanvas: HTMLCanvasElement;
	let chart: InstanceType<typeof Chart> | null = null;

	// 時間別データを準備
	$: chartData = (() => {
		const labels: string[] = [];
		const dailyProcessed: number[] = [];
		const monthlyProcessed: number[] = [];
		const dailyFailed: number[] = [];
		const monthlyFailed: number[] = [];

		hourlyMetrics.forEach((metric) => {
			const date = new Date(metric.timestamp * 1000);
			const hour = date.getHours();
			const day = date.getDate();
			labels.push(`${day}日 ${hour}時`);
			dailyProcessed.push(metric.dailySummariesProcessed);
			monthlyProcessed.push(metric.monthlySummariesProcessed);
			dailyFailed.push(metric.dailySummariesFailed);
			monthlyFailed.push(metric.monthlySummariesFailed);
		});

		return {
			labels,
			dailyProcessed,
			monthlyProcessed,
			dailyFailed,
			monthlyFailed,
		};
	})();

	// チャートを更新
	function updateChart() {
		if (!browser || !chart || !Chart) return;

		chart.data.labels = chartData.labels;
		chart.data.datasets[0].data = chartData.dailyProcessed;
		chart.data.datasets[1].data = chartData.monthlyProcessed;

		// ラベルとタイトルを更新
		if (chart.options.plugins?.title) {
			chart.options.plugins.title.text = $_("llm.metrics.processingActivity");
		}
		if (
			chart.options.scales?.x &&
			"title" in chart.options.scales.x &&
			chart.options.scales.x.title
		) {
			chart.options.scales.x.title.text = $_("llm.metrics.time");
		}
		if (
			chart.options.scales?.y &&
			"title" in chart.options.scales.y &&
			chart.options.scales.y.title
		) {
			chart.options.scales.y.title.text = $_("llm.metrics.processedCount");
		}
		chart.data.datasets[0].label = $_("llm.metrics.dailySummaries");
		chart.data.datasets[1].label = $_("llm.metrics.monthlySummaries");

		chart.update();
	}

	// チャートデータが変わったら更新
	$: if (chart && chartData) {
		updateChart();
	}

	// ロケールが変わったら更新
	$: if (chart && $locale) {
		updateChart();
	}

	// Chart.jsを動的にインポートする関数
	async function loadChartJS() {
		if (!browser) return;

		try {
			const chartModule = await import("chart.js");
			Chart = chartModule.Chart;
			CategoryScale = chartModule.CategoryScale;
			LinearScale = chartModule.LinearScale;
			PointElement = chartModule.PointElement;
			LineElement = chartModule.LineElement;
			LineController = chartModule.LineController;
			BarElement = chartModule.BarElement;
			BarController = chartModule.BarController;
			Title = chartModule.Title;
			Tooltip = chartModule.Tooltip;
			Legend = chartModule.Legend;

			Chart.register(
				CategoryScale,
				LinearScale,
				PointElement,
				LineElement,
				LineController,
				BarElement,
				BarController,
				Title,
				Tooltip,
				Legend,
			);
		} catch (error) {
			console.error("Failed to load Chart.js:", error);
		}
	}

	// チャートを作成する関数
	async function createChart() {
		if (!browser || !chartCanvas || !Chart) return;

		const ctx = chartCanvas.getContext("2d");
		if (!ctx) return;

		// 既存のチャートを破棄
		if (chart) {
			chart.destroy();
		}

		chart = new Chart(ctx, {
			type: "bar",
			data: {
				labels: chartData.labels,
				datasets: [
					{
						label: $_("llm.metrics.dailySummaries"),
						data: chartData.dailyProcessed,
						backgroundColor: "rgba(34, 197, 94, 0.6)",
						borderColor: "rgb(34, 197, 94)",
						borderWidth: 1,
					},
					{
						label: $_("llm.metrics.monthlySummaries"),
						data: chartData.monthlyProcessed,
						backgroundColor: "rgba(59, 130, 246, 0.6)",
						borderColor: "rgb(59, 130, 246)",
						borderWidth: 1,
					},
				],
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					title: {
						display: true,
						text: $_("llm.metrics.processingActivity"),
						font: {
							size: 16,
						},
					},
					legend: {
						display: true,
					},
					tooltip: {
						callbacks: {
							label: (context: import("chart.js").TooltipItem<"bar">) => {
								const value = context.parsed.y;
								return `${context.dataset.label}: ${value}`;
							},
						},
					},
				},
				scales: {
					x: {
						title: {
							display: true,
							text: $_("llm.metrics.time"),
						},
						grid: {
							display: true,
							color: "rgba(0, 0, 0, 0.1)",
						},
					},
					y: {
						title: {
							display: true,
							text: $_("llm.metrics.processedCount"),
						},
						beginAtZero: true,
						grid: {
							display: true,
							color: "rgba(0, 0, 0, 0.1)",
						},
					},
				},
				interaction: {
					intersect: false,
					mode: "index",
				},
			},
		});
	}

	// Chart.jsが読み込まれて翻訳が完了したらチャートを作成
	$: if (browser && !$isLoading && chartData && !chart && Chart) {
		createChart();
	}

	onMount(async () => {
		// Chart.jsを動的にロード
		await loadChartJS();

		// ホームからの遷移などではこっちが必要。翻訳は読み込み済みなので問題なし
		if (chartData && !chart && Chart) {
			createChart();
		}
	});

	onDestroy(() => {
		if (chart) {
			chart.destroy();
			chart = null;
		}
	});
</script>

<div
	class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6"
>
	<div class="h-64 md:h-80">
		<canvas bind:this={chartCanvas}></canvas>
	</div>
</div>