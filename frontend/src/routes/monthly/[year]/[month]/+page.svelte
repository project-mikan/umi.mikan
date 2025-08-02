<script lang="ts">
import { _ } from "svelte-i18n";
import { goto } from "$app/navigation";
import "$lib/i18n";
import { browser } from "$app/environment";
import type { DiaryEntry } from "$lib/grpc";
import type { PageData } from "./$types";

export let data: PageData;

let entries = data.entries;
let currentYear = data.year;
let currentMonth = data.month;
let _loading = false;

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

function _formatMonth(year: number, month: number): string {
	const date = new Date(year, month - 1, 1);
	return date.toLocaleDateString(undefined, { year: "numeric", month: "long" });
}

function _formatMonthOnly(month: number): string {
	const date = new Date(2000, month - 1, 1);
	return date.toLocaleDateString(undefined, { month: "long" });
}

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

function getWeekDays(): string[] {
	const days = [];
	const _date = new Date();
	// 日曜日から始まる週の各曜日を取得
	for (let i = 0; i < 7; i++) {
		const dayDate = new Date();
		dayDate.setDate(dayDate.getDate() - dayDate.getDay() + i);
		days.push(dayDate.toLocaleDateString(undefined, { weekday: "short" }));
	}
	return days;
}

const _weekDays = getWeekDays();

function _formatContentWithLineBreaks(content: string): string {
	return content.replace(/\n/g, "<br>");
}
</script>

<div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<!-- ヘッダー -->
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">
			{_formatMonth(currentYear, currentMonth)}
		</h1>
		<div class="flex space-x-4">
			<button
				on:click={_goToToday}
				class="bg-gray-600 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded"
			>
				{$_('monthly.thisMonth')}
			</button>
			<button
				on:click={() => goto('/')}
				class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
			>
				{$_('monthly.listView')}
			</button>
		</div>
	</div>

	<!-- 月ナビゲーション -->
	<div class="flex justify-center items-center mb-8 space-x-4">
		<button
			on:click={_previousMonth}
			class="p-2 rounded-full hover:bg-gray-100 transition-colors"
			aria-label={$_('monthly.previousMonth')}
		>
			<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
			</svg>
		</button>
		<h2 class="text-xl font-semibold text-gray-800 min-w-[200px] text-center">
			{_formatMonth(currentYear, currentMonth)}
			{#if _loading}
				<span class="ml-2 text-sm text-gray-500">読み込み中...</span>
			{/if}
		</h2>
		<button
			on:click={_nextMonth}
			class="p-2 rounded-full hover:bg-gray-100 transition-colors"
			aria-label={$_('monthly.nextMonth')}
		>
			<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
			</svg>
		</button>
	</div>

	<!-- カレンダー -->
	<div class="bg-white shadow rounded-lg overflow-hidden">
		<!-- 曜日ヘッダー -->
		<div class="grid grid-cols-7 bg-gray-50">
			{#each _weekDays as weekDay}
				<div class="p-4 text-center font-medium text-gray-700 border-r border-gray-200 last:border-r-0">
					{weekDay}
				</div>
			{/each}
		</div>

		<!-- カレンダーグリッド -->
		<div class="grid grid-cols-7">
			{#each calendarDays as day}
				<div class="h-32 border-r border-b border-gray-200 last:border-r-0">
					{#if day !== null}
						<button
							on:click={() => _navigateToEntry(day)}
							class="w-full h-full p-2 text-left hover:bg-gray-50 transition-colors cursor-pointer"
						>
							<div class="h-full flex flex-col">
								<!-- 日付 -->
								<div class="flex justify-between items-start mb-1">
									<span class="text-sm font-medium text-gray-700">{day}</span>
									{#if !entryMap.has(day)}
										<span class="text-xs text-blue-600 opacity-50">
											+
										</span>
									{/if}
								</div>

								<!-- 日記エントリ -->
								{#if entryMap.has(day)}
									{@const entry = entryMap.get(day)}
									<div class="flex-1 min-h-0">
										<div class="bg-blue-50 rounded p-2 h-full">
											<div class="text-xs text-blue-800 font-medium mb-1">
												{$_('monthly.entry')}
											</div>
											<div class="text-xs text-blue-600 line-clamp-2">
												{@html _formatContentWithLineBreaks(entry?.content ? entry.content.substring(0, 40) + (entry.content.length > 40 ? '...' : '') : '')}
											</div>
										</div>
									</div>
								{/if}
							</div>
						</button>
					{/if}
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.line-clamp-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>