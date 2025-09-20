<script lang="ts">
import { _, locale } from "svelte-i18n";
import { goto } from "$app/navigation";
import "$lib/i18n";
import { browser } from "$app/environment";
import { onMount } from "svelte";
import { authenticatedFetch } from "$lib/auth-client";
import type {
	DiaryEntry,
	GetDiaryEntriesByMonthResponse,
} from "$lib/grpc/diary/diary_pb";
import type { PageData } from "./$types";
import MonthlyCalendar from "$lib/components/molecules/MonthlyCalendar.svelte";
import MonthlyList from "$lib/components/molecules/MonthlyList.svelte";
import MonthYearSelector from "$lib/components/molecules/MonthYearSelector.svelte";
import CharacterCountChart from "$lib/components/molecules/CharacterCountChart.svelte";

interface MonthlySummary {
	id: string;
	month: { year: number; month: number };
	summary: string;
	createdAt: number;
	updatedAt: number;
}

export let data: PageData;

let entries: GetDiaryEntriesByMonthResponse | { entries: DiaryEntry[] } =
	data.entries;
let currentYear = data.year;
let currentMonth = data.month;
let _loading = false;
let showMonthSelector = false;
let summary: MonthlySummary | null = null;
let summaryLoading = false;
let showSummary = false;
let errorMessage = "";
let showErrorModal = false;
let hasNewerEntries = false;

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
		const response = await authenticatedFetch(
			`/api/diary/monthly/${year}/${month}`,
		);
		if (response.ok) {
			const newEntries:
				| GetDiaryEntriesByMonthResponse
				| { entries: DiaryEntry[] } = await response.json();
			entries = newEntries;
			currentYear = year;
			currentMonth = month;
		} else if (response.status === 401) {
			// Authentication failed completely, redirect to login
			console.warn("Authentication failed, redirecting to login");
			await goto("/login");
		} else {
			console.error(
				"Failed to fetch entries:",
				response.status,
				response.statusText,
			);
		}
	} catch (error) {
		console.error("Failed to fetch entries:", error);
		// If fetch fails completely, it might be a network issue or authentication problem
		// Don't redirect automatically in this case, just log the error
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

// サマリー関連の関数
async function fetchMonthlySummary() {
	if (!browser) return;

	try {
		const response = await authenticatedFetch(
			`/api/diary/summary/${currentYear}/${currentMonth}`,
		);
		if (response.ok) {
			const summaryData = await response.json();
			summary = summaryData;
		} else if (response.status === 404) {
			summary = null;
		} else if (response.status === 401) {
			await goto("/login");
		} else {
			console.error(
				"Failed to fetch summary:",
				response.status,
				response.statusText,
			);
		}
	} catch (error) {
		console.error("Failed to fetch summary:", error);
	}
}

async function generateMonthlySummary() {
	if (!browser) return;

	summaryLoading = true;
	try {
		const response = await authenticatedFetch(`/api/diary/summary/generate`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({
				year: currentYear,
				month: currentMonth,
			}),
		});

		if (response.ok) {
			const summaryData = await response.json();
			summary = summaryData;
			showSummary = true;
		} else if (response.status === 401) {
			await goto("/login");
		} else if (response.status === 404) {
			showError($_("monthly.summary.noEntries"));
		} else if (response.status === 400) {
			const errorData = await response.json();
			if (errorData.message?.includes("API key")) {
				showError($_("monthly.summary.noApiKey"));
			} else {
				showError($_("monthly.summary.error"));
			}
		} else {
			showError($_("monthly.summary.error"));
		}
	} catch (error) {
		console.error("Failed to generate summary:", error);
		showError($_("monthly.summary.error"));
	} finally {
		summaryLoading = false;
	}
}

// エラー表示用ヘルパー関数
function showError(message: string) {
	errorMessage = message;
	showErrorModal = true;
}

// 日記の最新更新日を取得
function getLatestEntryUpdate(): number {
	if (!entries || !entries.entries || entries.entries.length === 0) return 0;

	// 各日記エントリのupdatedAtフィールドから最も新しい更新日時を取得
	let latestUpdate = 0;
	for (const entry of entries.entries) {
		if (entry.updatedAt) {
			// 日記エントリは秒単位なのでミリ秒に変換
			const updatedAtMs = Number(entry.updatedAt) * 1000;
			if (updatedAtMs > latestUpdate) {
				latestUpdate = updatedAtMs;
			}
		}
	}

	return latestUpdate;
}

// サマリーが古いかどうかをチェック
function checkForNewerEntries() {
	if (!summary || !entries || !entries.entries) {
		hasNewerEntries = false;
		return;
	}

	const latestEntryTime = getLatestEntryUpdate(); // 既にミリ秒変換済み
	const summaryTime = Number(summary.updatedAt); // 既にミリ秒

	// サマリー更新後にエントリが追加/更新されているかチェック
	hasNewerEntries = latestEntryTime > summaryTime;
}

// エントリまたはサマリーが変わったときに更新検知を実行
$: if (entries || summary) {
	checkForNewerEntries();
}

// ローカルストレージのキーを生成
function getSummaryStorageKey(): string {
	return `summary-show-${currentYear}-${currentMonth}`;
}

// ページロード時の初期化処理
onMount(async () => {
	// 既存のサマリーがあるかチェック
	await fetchMonthlySummary();

	// サマリーが存在する場合のみ、前回の表示状態を復元
	if (summary) {
		const storageKey = getSummaryStorageKey();
		const storedShowState = localStorage.getItem(storageKey);
		if (storedShowState === "true") {
			showSummary = true;
		}
	}
});

// showSummaryの状態をローカルストレージに保存
$: if (browser && typeof window !== "undefined" && summary) {
	const storageKey = getSummaryStorageKey();
	localStorage.setItem(storageKey, showSummary.toString());
}

// 月が変わったときにサマリーをリセット
let previousYear = currentYear;
let previousMonth = currentMonth;

$: if (currentYear !== previousYear || currentMonth !== previousMonth) {
	// 以前の値を更新
	previousYear = currentYear;
	previousMonth = currentMonth;

	// 状態をリセット
	summary = null;
	showSummary = false;
	hasNewerEntries = false;

	// 新しい月のサマリーを取得（onMountで既に呼ばれている場合を除く）
	if (browser && (previousYear !== data.year || previousMonth !== data.month)) {
		fetchMonthlySummary().then(() => {
			// サマリーが存在する場合、前回の表示状態を復元
			if (summary) {
				const storageKey = getSummaryStorageKey();
				const storedShowState = localStorage.getItem(storageKey);
				if (storedShowState === "true") {
					showSummary = true;
				}
			}
		});
	}
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
		<div class="flex gap-2">
			{#if summary}
				<button
					on:click={() => showSummary = !showSummary}
					class="px-4 py-2 {showSummary ? 'bg-gray-600 hover:bg-gray-700' : 'bg-blue-600 hover:bg-blue-700'} text-white rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
				>
					{showSummary ? $_("monthly.summary.hide") : $_("monthly.summary.view")}
				</button>
				<button
					on:click={generateMonthlySummary}
					disabled={summaryLoading}
					class="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
				>
					{#if summaryLoading}
						{$_("monthly.summary.generating")}
					{:else}
						{$_("monthly.summary.regenerate")}
					{/if}
				</button>
			{:else}
				<button
					on:click={generateMonthlySummary}
					disabled={summaryLoading}
					class="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
				>
					{#if summaryLoading}
						{$_("monthly.summary.generating")}
					{:else}
						{$_("monthly.summary.generate")}
					{/if}
				</button>
			{/if}
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

	<!-- サマリー表示エリア -->
	{#if showSummary && summary}
		<div class="mb-8 bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
			<div class="p-6">
				<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
					{$_("diary.summary.label")}
				</h2>
				<div class="prose dark:prose-invert max-w-none">
					<p class="text-gray-700 dark:text-gray-300 whitespace-pre-wrap leading-relaxed">
						{summary.summary}
					</p>
				</div>
				{#if hasNewerEntries}
					<div class="mt-4 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-md">
						<p class="text-sm text-yellow-800 dark:text-yellow-200">
							💡 {$_("monthly.summary.updateAvailable")}
						</p>
					</div>
				{/if}
				<div class="mt-6 flex justify-between items-center text-sm text-gray-500 dark:text-gray-400">
					<span>
						{$_("common.createdAt")}: {new Date(summary.createdAt).toLocaleDateString()}
					</span>
					{#if summary.updatedAt !== summary.createdAt}
						<span>
							{$_("common.updatedAt")}: {new Date(summary.updatedAt).toLocaleDateString()}
						</span>
					{/if}
				</div>
			</div>
		</div>
	{/if}

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

	<!-- 日毎文字数グラフ -->
	<div class="mt-8">
		<CharacterCountChart
			{entryMap}
			year={currentYear}
			month={currentMonth}
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

<!-- Error Modal -->
{#if showErrorModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
		<div class="bg-white dark:bg-gray-800 rounded-lg max-w-md w-full">
			<div class="p-6">
				<div class="flex items-center mb-4">
					<div class="flex-shrink-0">
						<svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"></path>
						</svg>
					</div>
					<h3 class="ml-3 text-lg font-medium text-gray-900 dark:text-gray-100">
						{$_("common.error")}
					</h3>
				</div>
				<div class="mb-6">
					<p class="text-gray-700 dark:text-gray-300">
						{errorMessage}
					</p>
				</div>
				<div class="flex justify-end">
					<button
						on:click={() => showErrorModal = false}
						class="px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2"
					>
						{$_("common.close")}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

