<script lang="ts">
import { onMount, onDestroy } from "svelte";
import { browser } from "$app/environment";
import { _, locale, isLoading } from "svelte-i18n";
import "$lib/i18n";
import type { DiaryEntry } from "$lib/grpc/diary/diary_pb";

// Chart.js の型定義
let Chart: typeof import("chart.js").Chart;
let CategoryScale: typeof import("chart.js").CategoryScale;
let LinearScale: typeof import("chart.js").LinearScale;
let PointElement: typeof import("chart.js").PointElement;
let LineElement: typeof import("chart.js").LineElement;
let LineController: typeof import("chart.js").LineController;
let Title: typeof import("chart.js").Title;
let Tooltip: typeof import("chart.js").Tooltip;
let Legend: typeof import("chart.js").Legend;
let Filler: typeof import("chart.js").Filler;

export let entryMap: Map<number, DiaryEntry>;
export let year: number;
export let month: number;

let chartCanvas: HTMLCanvasElement;
let chart: InstanceType<typeof Chart> | null = null;

// 指定された月の日数を取得
function getDaysInMonth(year: number, month: number): number {
	return new Date(year, month, 0).getDate();
}

// 文字数データを準備
$: chartData = (() => {
	const daysInMonth = getDaysInMonth(year, month);
	const labels: string[] = [];
	const data: number[] = [];

	for (let day = 1; day <= daysInMonth; day++) {
		labels.push(day.toString());
		const entry = entryMap.get(day);
		const characterCount = entry?.content ? entry.content.length : 0;
		data.push(characterCount);
	}

	return { labels, data };
})();

// チャートを更新
function updateChart() {
	if (!browser || !chart || !Chart) return;

	chart.data.labels = chartData.labels;
	chart.data.datasets[0].data = chartData.data;

	// ラベルとタイトルを更新
	if (chart.options.plugins?.title) {
		chart.options.plugins.title.text = $_("chart.dailyCharacterCount");
	}
	if (
		chart.options.scales?.x &&
		"title" in chart.options.scales.x &&
		chart.options.scales.x.title
	) {
		chart.options.scales.x.title.text = $_("chart.day");
	}
	if (
		chart.options.scales?.y &&
		"title" in chart.options.scales.y &&
		chart.options.scales.y.title
	) {
		chart.options.scales.y.title.text = $_("chart.characterCount");
	}
	chart.data.datasets[0].label = $_("chart.characterCount");

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
		Title = chartModule.Title;
		Tooltip = chartModule.Tooltip;
		Legend = chartModule.Legend;
		Filler = chartModule.Filler;

		Chart.register(
			CategoryScale,
			LinearScale,
			PointElement,
			LineElement,
			LineController,
			Title,
			Tooltip,
			Legend,
			Filler,
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
		type: "line",
		data: {
			labels: chartData.labels,
			datasets: [
				{
					label: $_("chart.characterCount"),
					data: chartData.data,
					borderColor: "rgb(56, 189, 248)",
					backgroundColor: "rgba(56, 189, 248, 0.1)",
					borderWidth: 2,
					fill: true,
					tension: 0.1,
					pointBackgroundColor: "rgb(56, 189, 248)",
					pointBorderColor: "rgb(56, 189, 248)",
					pointHoverBackgroundColor: "rgb(14, 165, 233)",
					pointHoverBorderColor: "rgb(14, 165, 233)",
				},
			],
		},
		options: {
			responsive: true,
			maintainAspectRatio: false,
			plugins: {
				title: {
					display: true,
					text: $_("chart.dailyCharacterCount"),
					font: {
						size: 16,
					},
				},
				legend: {
					display: false,
				},
				tooltip: {
					callbacks: {
						label: (context: import("chart.js").TooltipItem<"line">) => {
							const value = context.parsed.y;
							return `${$_("chart.characterCount")}: ${value}${$_("chart.charactersUnit")}`;
						},
					},
				},
			},
			scales: {
				x: {
					title: {
						display: true,
						text: $_("chart.day"),
					},
					grid: {
						display: true,
						color: "rgba(0, 0, 0, 0.1)",
					},
				},
				y: {
					title: {
						display: true,
						text: $_("chart.characterCount"),
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
