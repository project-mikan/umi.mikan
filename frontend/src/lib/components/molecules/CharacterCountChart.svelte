<script lang="ts">
import { onMount, onDestroy } from "svelte";
import { _, locale, isLoading } from "svelte-i18n";
import "$lib/i18n";
import {
	Chart,
	CategoryScale,
	LinearScale,
	PointElement,
	LineElement,
	LineController,
	Title,
	Tooltip,
	Legend,
	Filler,
} from "chart.js";
import type { DiaryEntry } from "$lib/grpc/diary/diary_pb";

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

export let entryMap: Map<number, DiaryEntry>;
export let year: number;
export let month: number;

let chartCanvas: HTMLCanvasElement;
let chart: Chart | null = null;

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
	if (!chart) return;

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

// チャートを作成する関数
function createChart() {
	if (!chartCanvas) return;

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
						label: (context) => {
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

// 翻訳が読み込まれたら初回チャートを作成
$: if (!$isLoading && chartData && !chart) {
	createChart();
}

onMount(() => {
	// onMountでは何もしない（リアクティブ文で処理）
	// ここで作ると翻訳データが使えずchart.dayみたいな値になるので
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

