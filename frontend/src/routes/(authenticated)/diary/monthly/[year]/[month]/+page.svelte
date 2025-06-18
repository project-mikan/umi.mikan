<script lang="ts">
import { goto } from "$app/navigation";
import type { PageData } from "./$types";
import type { DiaryEntry } from "$lib/grpc";

export let data: PageData;

function formatMonth(year: number, month: number): string {
	const date = new Date(year, month - 1, 1);
	return date.toLocaleDateString(undefined, { year: 'numeric', month: 'long' });
}

function formatMonthOnly(month: number): string {
	const date = new Date(2000, month - 1, 1);
	return date.toLocaleDateString(undefined, { month: 'long' });
}

function getDaysInMonth(year: number, month: number): number {
	return new Date(year, month, 0).getDate();
}

function getFirstDayOfWeek(year: number, month: number): number {
	return new Date(year, month - 1, 1).getDay();
}

function createEntry(day: number) {
	const dateStr = `${data.year}-${String(data.month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
	goto(`/diary/create?date=${dateStr}`);
}

function viewEntry(entry: DiaryEntry) {
	const date = entry.date;
	if (date) {
		const dateStr = `${date.year}-${String(date.month).padStart(2, '0')}-${String(date.day).padStart(2, '0')}`;
		goto(`/diary/${dateStr}`);
	}
}

function editEntry(entry: DiaryEntry) {
	const date = entry.date;
	if (date) {
		const dateStr = `${date.year}-${String(date.month).padStart(2, '0')}-${String(date.day).padStart(2, '0')}`;
		goto(`/diary/edit/${dateStr}`);
	}
}

function previousMonth() {
	const prevMonth = data.month === 1 ? 12 : data.month - 1;
	const prevYear = data.month === 1 ? data.year - 1 : data.year;
	goto(`/diary/monthly/${prevYear}/${prevMonth}`);
}

function nextMonth() {
	const nextMonth = data.month === 12 ? 1 : data.month + 1;
	const nextYear = data.month === 12 ? data.year + 1 : data.year;
	goto(`/diary/monthly/${nextYear}/${nextMonth}`);
}

function goToToday() {
	const now = new Date();
	goto(`/diary/monthly/${now.getFullYear()}/${now.getMonth() + 1}`);
}

// カレンダーデータの準備
const daysInMonth = getDaysInMonth(data.year, data.month);
const firstDayOfWeek = getFirstDayOfWeek(data.year, data.month);
const calendarDays: (number | null)[] = [];

// 月の最初の日までの空白
for (let i = 0; i < firstDayOfWeek; i++) {
	calendarDays.push(null);
}

// 月の日付
for (let day = 1; day <= daysInMonth; day++) {
	calendarDays.push(day);
}

// 日記エントリをマップに変換
const entryMap = new Map<number, DiaryEntry>();
if (data.entries.entries) {
	for (const entry of data.entries.entries) {
		if (entry.date) {
			entryMap.set(entry.date.day, entry);
		}
	}
}

function getWeekDays(): string[] {
	const days = [];
	const date = new Date();
	// 日曜日から始まる週の各曜日を取得
	for (let i = 0; i < 7; i++) {
		const dayDate = new Date();
		dayDate.setDate(dayDate.getDate() - dayDate.getDay() + i);
		days.push(dayDate.toLocaleDateString(undefined, { weekday: 'short' }));
	}
	return days;
}

const weekDays = getWeekDays();
</script>

<div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<!-- ヘッダー -->
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">
			{formatMonth(data.year, data.month)}
		</h1>
		<div class="flex space-x-4">
			<button
				on:click={goToToday}
				class="bg-gray-600 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded"
			>
				今月
			</button>
			<button
				on:click={() => goto('/diary')}
				class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
			>
				一覧表示
			</button>
		</div>
	</div>

	<!-- 月ナビゲーション -->
	<div class="flex justify-center items-center mb-8 space-x-4">
		<button
			on:click={previousMonth}
			class="p-2 rounded-full hover:bg-gray-100 transition-colors"
		>
			<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
			</svg>
		</button>
		<h2 class="text-xl font-semibold text-gray-800 min-w-[200px] text-center">
			{formatMonth(data.year, data.month)}
		</h2>
		<button
			on:click={nextMonth}
			class="p-2 rounded-full hover:bg-gray-100 transition-colors"
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
			{#each weekDays as weekDay}
				<div class="p-4 text-center font-medium text-gray-700 border-r border-gray-200 last:border-r-0">
					{weekDay}
				</div>
			{/each}
		</div>

		<!-- カレンダーグリッド -->
		<div class="grid grid-cols-7">
			{#each calendarDays as day}
				<div class="h-32 border-r border-b border-gray-200 last:border-r-0 p-2">
					{#if day !== null}
						<div class="h-full flex flex-col">
							<!-- 日付 -->
							<div class="flex justify-between items-start mb-1">
								<span class="text-sm font-medium text-gray-700">{day}</span>
								{#if !entryMap.has(day)}
									<button
										on:click={() => createEntry(day)}
										class="text-xs text-blue-600 hover:text-blue-800 opacity-50 hover:opacity-100"
										title="日記を書く"
									>
										+
									</button>
								{/if}
							</div>

							<!-- 日記エントリ -->
							{#if entryMap.has(day)}
								{@const entry = entryMap.get(day)}
								<div class="flex-1 min-h-0">
									<div class="bg-blue-50 rounded p-2 h-full hover:bg-blue-100 cursor-pointer transition-colors">
										<div class="text-xs text-blue-800 font-medium mb-1">
											{entry?.title || '無題'}
										</div>
										<div class="text-xs text-blue-600 line-clamp-2">
											{entry?.content ? entry.content.substring(0, 40) + (entry.content.length > 40 ? '...' : '') : ''}
										</div>
										<div class="flex justify-end space-x-1 mt-1">
											<button
												on:click={() => viewEntry(entry)}
												class="text-xs text-blue-600 hover:text-blue-800"
											>
												詳細
											</button>
											<button
												on:click={() => editEntry(entry)}
												class="text-xs text-green-600 hover:text-green-800"
											>
												編集
											</button>
										</div>
									</div>
								</div>
							{/if}
						</div>
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