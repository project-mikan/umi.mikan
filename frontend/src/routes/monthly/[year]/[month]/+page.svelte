<script lang="ts">
import { _, locale } from "svelte-i18n";
import { goto } from "$app/navigation";
import "$lib/i18n";
import { browser } from "$app/environment";
import type { DiaryEntry } from "$lib/grpc";
import type { PageData } from "./$types";
import MonthlyCalendar from "$lib/components/molecules/MonthlyCalendar.svelte";
import MonthlyList from "$lib/components/molecules/MonthlyList.svelte";
import MonthYearSelector from "$lib/components/molecules/MonthYearSelector.svelte";

export let data: PageData;

let entries = data.entries;
let currentYear = data.year;
let currentMonth = data.month;
let _loading = false;
let showMonthSelector = false;

// データの更新
$: {
	entries = data.entries;
	currentYear = data.year;
	currentMonth = data.month;
}

// クライアントサイドでデータを再取得する関数
async function fetchMonthData(year: number, month: number) {
	if (!browser) return;

	_loading = true;
	try {
		const response = await fetch(`/api/diary/monthly/${year}/${month}`);
		if (response.ok) {
			const newEntries = await response.json();
			entries = newEntries;
			currentYear = year;
			currentMonth = month;
		}
	} catch (error) {
		console.error("Failed to fetch entries:", error);
	} finally {
		_loading = false;
	}
}

// Reactive date formatting functions
$: _formatMonth = (year: number, month: number): string => {
	const date = new Date(year, month - 1, 1);
	return date.toLocaleDateString($locale || "en", {
		year: "numeric",
		month: "long",
	});
};

$: _formatMonthOnly = (month: number): string => {
	const date = new Date(2000, month - 1, 1);
	return date.toLocaleDateString($locale || "en", { month: "long" });
};

function getDaysInMonth(year: number, month: number): number {
	return new Date(year, month, 0).getDate();
}

function getFirstDayOfWeek(year: number, month: number): number {
	return new Date(year, month - 1, 1).getDay();
}

function _createEntry(day: number) {
	const dateStr = `${currentYear}-${String(currentMonth).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
	goto(`/${dateStr}`);
}

function _navigateToEntry(day: number) {
	const dateStr = `${currentYear}-${String(currentMonth).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
	goto(`/${dateStr}`);
}

async function _previousMonth() {
	const prevMonth = currentMonth === 1 ? 12 : currentMonth - 1;
	const prevYear = currentMonth === 1 ? currentYear - 1 : currentYear;
	await fetchMonthData(prevYear, prevMonth);
	await goto(`/monthly/${prevYear}/${prevMonth}`, { replaceState: true });
}

async function _nextMonth() {
	const nextMonth = currentMonth === 12 ? 1 : currentMonth + 1;
	const nextYear = currentMonth === 12 ? currentYear + 1 : currentYear;
	await fetchMonthData(nextYear, nextMonth);
	await goto(`/monthly/${nextYear}/${nextMonth}`, { replaceState: true });
}

async function _goToToday() {
	const now = new Date();
	const year = now.getFullYear();
	const month = now.getMonth() + 1;
	await fetchMonthData(year, month);
	await goto(`/monthly/${year}/${month}`, { replaceState: true });
}

function _showMonthSelector() {
	showMonthSelector = true;
}

async function _handleMonthSelect(year: number, month: number) {
	showMonthSelector = false;
	await fetchMonthData(year, month);
	await goto(`/monthly/${year}/${month}`, { replaceState: true });
}

function _handleMonthSelectorCancel() {
	showMonthSelector = false;
}

// カレンダーデータの準備（リアクティブ）
$: daysInMonth = getDaysInMonth(currentYear, currentMonth);
$: firstDayOfWeek = getFirstDayOfWeek(currentYear, currentMonth);
$: calendarDays = (() => {
	const days: (number | null)[] = [];
	// 月の最初の日までの空白
	for (let i = 0; i < firstDayOfWeek; i++) {
		days.push(null);
	}
	// 月の日付
	for (let day = 1; day <= daysInMonth; day++) {
		days.push(day);
	}
	return days;
})();

// 日記エントリをマップに変換（リアクティブ）
$: entryMap = (() => {
	const map = new Map<number, DiaryEntry>();
	if (entries && Array.isArray(entries.entries)) {
		for (const entry of entries.entries) {
			if (entry?.date) {
				map.set(entry.date.day, entry);
			}
		}
	}
	return map;
})();

// Reactive weekdays
$: _weekDays = (() => {
	const days = [];
	const _date = new Date();
	// 日曜日から始まる週の各曜日を取得
	for (let i = 0; i < 7; i++) {
		const dayDate = new Date();
		dayDate.setDate(dayDate.getDate() - dayDate.getDay() + i);
		days.push(
			dayDate.toLocaleDateString($locale || "en", { weekday: "short" }),
		);
	}
	return days;
})();
</script>

<div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<!-- ヘッダー -->
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
			{_formatMonth(currentYear, currentMonth)}
		</h1>
		<div class="flex space-x-4">
			<button
				on:click={_goToToday}
				class="bg-gray-600 hover:bg-gray-700 dark:bg-gray-500 dark:hover:bg-gray-600 text-white font-bold py-2 px-4 rounded"
			>
				{$_("monthly.thisMonth")}
			</button>
			<button
				on:click={() => goto("/")}
				class="bg-blue-600 hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 text-white font-bold py-2 px-4 rounded"
			>
				{$_("home.name")}
			</button>
		</div>
	</div>

	<!-- 月ナビゲーション -->
	<div class="flex justify-center items-center mb-8 space-x-4">
		<button
			on:click={_previousMonth}
			class="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-gray-700 dark:text-gray-300"
			aria-label={$_("monthly.previousMonth")}
		>
			<svg
				class="w-6 h-6"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M15 19l-7-7 7-7"
				></path>
			</svg>
		</button>
		<button 
			on:click={_showMonthSelector}
			class="text-xl font-semibold text-white bg-green-600 hover:bg-green-700 dark:bg-green-500 dark:hover:bg-green-600 min-w-[200px] text-center rounded-md px-4 py-2 transition-colors cursor-pointer shadow-md hover:shadow-lg focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
		>
			{_formatMonth(currentYear, currentMonth)}
			{#if _loading}
				<span class="ml-2 text-sm text-green-200 dark:text-green-300">読み込み中...</span>
			{/if}
		</button>
		<button
			on:click={_nextMonth}
			class="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-gray-700 dark:text-gray-300"
			aria-label={$_("monthly.nextMonth")}
		>
			<svg
				class="w-6 h-6"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M9 5l7 7-7 7"
				></path>
			</svg>
		</button>
	</div>

	<!-- デスクトップ・タブレット: カレンダー表示 -->
	<div class="hidden md:block">
		<MonthlyCalendar
			{calendarDays}
			{entryMap}
			weekDays={_weekDays}
			onNavigateToEntry={_navigateToEntry}
		/>
	</div>

	<!-- モバイル: リスト表示 -->
	<div class="block md:hidden">
		<MonthlyList
			{daysInMonth}
			{currentYear}
			{currentMonth}
			{entryMap}
			onNavigateToEntry={_navigateToEntry}
		/>
	</div>
</div>

<!-- Month/Year Selector Modal -->
<MonthYearSelector
	isOpen={showMonthSelector}
	{currentYear}
	{currentMonth}
	onSelect={_handleMonthSelect}
	onCancel={_handleMonthSelectorCancel}
/>

