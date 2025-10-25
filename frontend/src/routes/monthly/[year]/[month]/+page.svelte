<script lang="ts">
	import { _, locale } from "svelte-i18n";
	import { goto } from "$app/navigation";
	import "$lib/i18n";
	import { browser } from "$app/environment";
	import { onMount } from "svelte";
	import { authenticatedFetch } from "$lib/auth-client";
	import type {
		DiaryEntry,
		DiaryEntityOutput,
		GetDiaryEntriesByMonthResponse,
		YMD,
	} from "$lib/grpc/diary/diary_pb";
	import type { PageData } from "./$types";
	import MonthlyCalendar from "$lib/components/molecules/MonthlyCalendar.svelte";
	import MonthlyList from "$lib/components/molecules/MonthlyList.svelte";
	import MonthYearSelector from "$lib/components/molecules/MonthYearSelector.svelte";
	import CharacterCountChart from "$lib/components/molecules/CharacterCountChart.svelte";
	import SummaryDisplay from "$lib/components/molecules/SummaryDisplay.svelte";

	$: title = $_("page.title.calendar");

	interface MonthlySummary {
		id: string;
		month: { year: number; month: number };
		summary: string;
		createdAt: number;
		updatedAt: number;
	}

	interface SerializedDiaryEntry {
		id: string;
		date?: { year: number; month: number; day: number };
		content: string;
		createdAt: number;
		updatedAt: number;
		diaryEntities?: DiaryEntityOutput[];
	}

	interface SerializedGetDiaryEntriesByMonthResponse {
		entries: SerializedDiaryEntry[];
	}

	export let data: PageData;

	let entries: SerializedGetDiaryEntriesByMonthResponse = data.entries;
	let currentYear = data.year;
	let currentMonth = data.month;
	let _loading = false;
	let showMonthSelector = false;
	let errorMessage = "";
	let showErrorModal = false;
	let summaryError: string | null = null;
	let isCurrentMonth = false;
	let isFutureMonth = false;
	let hasEntries = false;
	let isSummaryGenerating = false; // 要約生成中のフラグ
	let monthlySummary: MonthlySummary | null = null;
	let isMonthlySummaryOutdated = false;
	let lastMonthlySummaryUpdateTime = 0; // 最後に月次要約が更新された時刻（ミリ秒）
	let isInitialLoad = true; // 初回読み込みかどうかのフラグ

	// Check if user has LLM key configured
	$: existingLLMKey = data.user?.llmKeys?.find((key) => key.llmProvider === 1);
	$: hasLLMKey = !!existingLLMKey;

	// 現在の月かどうかの判定（リアクティブ）
	$: {
		const now = new Date();
		const currentDate = new Date(currentYear, currentMonth - 1, 1);
		const todayDate = new Date(now.getFullYear(), now.getMonth(), 1);

		isCurrentMonth = currentDate.getTime() === todayDate.getTime();
		isFutureMonth = currentDate.getTime() > todayDate.getTime();
	}

	// 日記エントリがあるかどうかの判定（リアクティブ）
	$: {
		hasEntries = entries?.entries && entries.entries.length > 0;
	}

	// データの更新
	$: {
		entries = data.entries;
		currentYear = data.year;
		currentMonth = data.month;
		// 月が変わった時は初回読み込み扱いにリセット
		isInitialLoad = true;
	}

	// 月次要約が古いかどうかを判定（リアクティブ）
	$: isMonthlySummaryOutdated = (() => {
		if (!monthlySummary || !entries?.entries) return false;

		// その月の全ての日記エントリの最新更新日時を取得
		const latestEntryUpdatedAt = entries.entries.reduce((latest, entry) => {
			const entryUpdatedAt = Number(entry.updatedAt) * 1000; // 秒 → ミリ秒
			return entryUpdatedAt > latest ? entryUpdatedAt : latest;
		}, 0);

		// 月次要約の更新日時（既にミリ秒）
		const summaryUpdatedAt = Number(monthlySummary.updatedAt);

		// 要約が最近更新された場合（5秒以内）は古くないとみなす
		const now = Date.now();
		const recentlyUpdated =
			lastMonthlySummaryUpdateTime > 0 &&
			now - lastMonthlySummaryUpdateTime < 5000;

		// 要約が最新の日記エントリよりも新しい場合、または最近更新された場合は古くない
		const isOutdated =
			latestEntryUpdatedAt > summaryUpdatedAt && !recentlyUpdated;

		return isOutdated;
	})();

	// クライアントサイドでデータを再取得する関数
	async function fetchMonthData(year: number, month: number) {
		if (!browser) return;

		_loading = true;
		try {
			const response = await authenticatedFetch(
				`/api/diary/monthly/${year}/${month}`,
			);
			if (response.ok) {
				const newEntries: SerializedGetDiaryEntriesByMonthResponse =
					await response.json();
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

	function handleSummaryUpdated(event: CustomEvent) {
		const newSummary = event.detail.summary;
		const oldSummary = monthlySummary;

		// 要約が実際に変更されたかどうかを確認
		// 初回読み込み時は変更とみなさない
		const actuallyUpdated =
			!isInitialLoad &&
			oldSummary &&
			(oldSummary.updatedAt !== newSummary.updatedAt ||
				oldSummary.summary !== newSummary.summary);

		// 常に要約は更新するが、初回読み込み時は時刻は記録しない
		monthlySummary = newSummary;
		if (actuallyUpdated) {
			lastMonthlySummaryUpdateTime = Date.now();
		}

		// 初回読み込み完了をマーク
		if (isInitialLoad) {
			isInitialLoad = false;
		}
	}

	function handleSummaryError(event: CustomEvent) {
		summaryError = event.detail.message;
	}

	function handleGenerationStarted() {
		isSummaryGenerating = true;
		summaryError = null;
	}

	function handleGenerationCompleted() {
		isSummaryGenerating = false;
	}

	// エラー表示用ヘルパー関数
	function showError(message: string) {
		errorMessage = message;
		showErrorModal = true;
	}

	// 無効化メッセージを取得（リアクティブ）
	$: getDisabledMessage = (): string => {
		if (isFutureMonth) {
			return $_("monthly.summary.futureMonthError");
		}
		if (isCurrentMonth) {
			return $_("monthly.summary.currentMonthError");
		}
		if (!hasEntries) {
			return $_("monthly.summary.noEntriesError");
		}
		return "";
	};

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
					// シリアライズされたエントリをDiaryEntry形式に変換
					const compatibleEntry: DiaryEntry = {
						id: entry.id,
						content: entry.content,
						createdAt: BigInt(entry.createdAt),
						updatedAt: BigInt(entry.updatedAt),
						date: {
							year: entry.date.year,
							month: entry.date.month,
							day: entry.date.day,
							$typeName: "diary.YMD" as const,
						} as YMD,
						diaryEntities: entry.diaryEntities || [],
						$typeName: "diary.DiaryEntry" as const,
					};
					map.set(entry.date.day, compatibleEntry);
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

<svelte:head>
	<title>{title}</title>
</svelte:head>

<div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<!-- 月ナビゲーション（最上段に配置） -->
	<div class="flex justify-center items-center mb-6 space-x-4">
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

	<!-- ヘッダー -->
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
			{_formatMonth(currentYear, currentMonth)}
		</h1>
	</div>

	<!-- サマリー表示エリア -->
	<SummaryDisplay
		type="monthly"
		fetchUrl="/api/diary/summary/{currentYear}/{currentMonth}"
		generateUrl="/api/diary/summary/generate"
		generatePayload={{
			year: currentYear,
			month: currentMonth
		}}
		isDisabled={isFutureMonth || isCurrentMonth || !hasEntries}
		disabledMessage={getDisabledMessage()}
		{hasLLMKey}
		isSummaryOutdated={isMonthlySummaryOutdated}
		isGenerating={isSummaryGenerating}
		on:summaryUpdated={handleSummaryUpdated}
		on:summaryError={handleSummaryError}
		on:generationStarted={handleGenerationStarted}
		on:generationCompleted={handleGenerationCompleted}
	/>

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


