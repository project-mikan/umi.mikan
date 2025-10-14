<script lang="ts">
import { _ } from "svelte-i18n";
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import { beforeNavigate } from "$app/navigation";
import { onMount } from "svelte";
import "$lib/i18n";
import Button from "$lib/components/atoms/Button.svelte";
import SaveButton from "$lib/components/atoms/SaveButton.svelte";
import DiaryCard from "$lib/components/molecules/DiaryCard.svelte";
import FormField from "$lib/components/molecules/FormField.svelte";
import TimeProgressBar from "$lib/components/molecules/TimeProgressBar.svelte";
import PWAInstallButton from "$lib/components/PWAInstallButton.svelte";
import { createSubmitHandler } from "$lib/utils/form-utils";
import type { DiaryEntry, YMD } from "$lib/grpc/diary/diary_pb";
import type { PageData } from "./$types";

$: title = $_("page.title.home");

export let data: PageData;

// dataが更新されたときに自動的に更新されるようにリアクティブ宣言を使用
$: todayContent = data.today.entry?.content || "";
$: yesterdayContent = data.yesterday.entry?.content || "";
$: dayBeforeYesterdayContent = data.dayBeforeYesterday.entry?.content || "";

let formElement: HTMLFormElement;
let yesterdayFormElement: HTMLFormElement;
let dayBeforeYesterdayFormElement: HTMLFormElement;
let [todayLoading, yesterdayLoading, dayBeforeLoading] = [false, false, false];
let [todaySaved, yesterdaySaved, dayBeforeSaved] = [false, false, false];

// 明示的に選択されたエンティティの情報
let todaySelectedEntities: {
	entityId: string;
	positions: { start: number; end: number }[];
}[] = [];
let yesterdaySelectedEntities: {
	entityId: string;
	positions: { start: number; end: number }[];
}[] = [];
let dayBeforeYesterdaySelectedEntities: {
	entityId: string;
	positions: { start: number; end: number }[];
}[] = [];

// Character count calculations
$: todayCharacterCount = todayContent ? todayContent.length : 0;
$: yesterdayCharacterCount = yesterdayContent ? yesterdayContent.length : 0;
$: dayBeforeYesterdayCharacterCount = dayBeforeYesterdayContent
	? dayBeforeYesterdayContent.length
	: 0;

// 未保存状態の管理
let initialTodayContent = data.today.entry?.content || "";
let initialYesterdayContent = data.yesterday.entry?.content || "";
let initialDayBeforeYesterdayContent =
	data.dayBeforeYesterday.entry?.content || "";
let allowNavigation = false;

// 各日記の未保存状態を監視
$: todayHasUnsavedChanges =
	todayContent !== initialTodayContent && !allowNavigation;
$: yesterdayHasUnsavedChanges =
	yesterdayContent !== initialYesterdayContent && !allowNavigation;
$: dayBeforeYesterdayHasUnsavedChanges =
	dayBeforeYesterdayContent !== initialDayBeforeYesterdayContent &&
	!allowNavigation;

// いずれか1つでも未保存の変更があるかチェック
$: hasAnyUnsavedChanges =
	todayHasUnsavedChanges ||
	yesterdayHasUnsavedChanges ||
	dayBeforeYesterdayHasUnsavedChanges;

// データが変更された時に初期コンテンツをリセット
$: {
	initialTodayContent = data.today.entry?.content || "";
	initialYesterdayContent = data.yesterday.entry?.content || "";
	initialDayBeforeYesterdayContent =
		data.dayBeforeYesterday.entry?.content || "";
	todayContent = initialTodayContent;
	yesterdayContent = initialYesterdayContent;
	dayBeforeYesterdayContent = initialDayBeforeYesterdayContent;
	allowNavigation = false;
}

function getMonthlyUrl(): string {
	const now = new Date();
	return `/monthly/${now.getFullYear()}/${now.getMonth() + 1}`;
}

function formatDateStr(ymd: YMD): string {
	return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
}

function viewEntry(entry: DiaryEntry) {
	const date = entry.date;
	if (date) {
		const dateStr = formatDateStr(date);
		goto(`/${dateStr}`);
	}
}

function handleSave() {
	formElement?.requestSubmit();
}

function handleYesterdaySave() {
	yesterdayFormElement?.requestSubmit();
}

function handleDayBeforeYesterdaySave() {
	dayBeforeYesterdayFormElement?.requestSubmit();
}

// ページ遷移前の警告
beforeNavigate((navigation) => {
	if (hasAnyUnsavedChanges && !allowNavigation) {
		if (!confirm($_("diary.unsavedChangesWarning"))) {
			navigation.cancel();
		}
	}
});

// スクロール位置に基づいて表示する保存ボタンを決定
let activeSection: "today" | "yesterday" | "dayBeforeYesterday" = "today";
let todayCard: HTMLElement;
let yesterdayCard: HTMLElement;
let dayBeforeYesterdayCard: HTMLElement;

// ブラウザのページ離脱時の警告
onMount(() => {
	const handleBeforeUnload = (e: BeforeUnloadEvent) => {
		if (hasAnyUnsavedChanges) {
			e.preventDefault();
			e.returnValue = "";
		}
	};

	// スクロール位置を監視して、現在表示中のセクションを判定
	const handleScroll = () => {
		const scrollY = window.scrollY;
		const viewportHeight = window.innerHeight;
		const viewportCenter = scrollY + viewportHeight / 2;

		// 各カードの位置を取得
		const todayRect = todayCard?.getBoundingClientRect();
		const yesterdayRect = yesterdayCard?.getBoundingClientRect();
		const dayBeforeYesterdayRect =
			dayBeforeYesterdayCard?.getBoundingClientRect();

		// 画面中央を含むカードを優先的に選択
		// 画面中央がカードの範囲内にある場合、そのカードを選択
		if (
			todayRect &&
			todayRect.top + scrollY <= viewportCenter &&
			todayRect.bottom + scrollY >= viewportCenter
		) {
			activeSection = "today";
			return;
		}

		if (
			yesterdayRect &&
			yesterdayRect.top + scrollY <= viewportCenter &&
			yesterdayRect.bottom + scrollY >= viewportCenter
		) {
			activeSection = "yesterday";
			return;
		}

		if (
			dayBeforeYesterdayRect &&
			dayBeforeYesterdayRect.top + scrollY <= viewportCenter &&
			dayBeforeYesterdayRect.bottom + scrollY >= viewportCenter
		) {
			activeSection = "dayBeforeYesterday";
			return;
		}

		// 画面中央がどのカードにも含まれない場合、画面内で最も近いカードを選択
		const candidates: Array<{
			section: "today" | "yesterday" | "dayBeforeYesterday";
			distance: number;
		}> = [];

		if (todayRect && todayRect.top < viewportHeight && todayRect.bottom > 0) {
			const todayCenter = todayRect.top + scrollY + todayRect.height / 2;
			const todayDistance = Math.abs(viewportCenter - todayCenter);
			candidates.push({ section: "today", distance: todayDistance });
		}

		if (
			yesterdayRect &&
			yesterdayRect.top < viewportHeight &&
			yesterdayRect.bottom > 0
		) {
			const yesterdayCenter =
				yesterdayRect.top + scrollY + yesterdayRect.height / 2;
			const yesterdayDistance = Math.abs(viewportCenter - yesterdayCenter);
			candidates.push({ section: "yesterday", distance: yesterdayDistance });
		}

		if (
			dayBeforeYesterdayRect &&
			dayBeforeYesterdayRect.top < viewportHeight &&
			dayBeforeYesterdayRect.bottom > 0
		) {
			const dayBeforeYesterdayCenter =
				dayBeforeYesterdayRect.top +
				scrollY +
				dayBeforeYesterdayRect.height / 2;
			const dayBeforeYesterdayDistance = Math.abs(
				viewportCenter - dayBeforeYesterdayCenter,
			);
			candidates.push({
				section: "dayBeforeYesterday",
				distance: dayBeforeYesterdayDistance,
			});
		}

		if (candidates.length > 0) {
			candidates.sort((a, b) => a.distance - b.distance);
			activeSection = candidates[0].section;
		}
	};

	window.addEventListener("beforeunload", handleBeforeUnload);
	window.addEventListener("scroll", handleScroll);

	// 初期表示時に一度実行
	handleScroll();

	return () => {
		window.removeEventListener("beforeunload", handleBeforeUnload);
		window.removeEventListener("scroll", handleScroll);
	};
});
</script>

<svelte:head>
	<title>{title}</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">{$_("diary.title")}</h1>
	</div>

	<div class="mb-8">
		<TimeProgressBar />
	</div>

	<div class="space-y-6">
		<div bind:this={todayCard}>
		<DiaryCard
			title={$_("diary.today")}
			entry={data.today.entry}
			showForm={true}
			href={`/${formatDateStr(data.today.date)}`}
		>
			<form
				bind:this={formElement}
				method="POST"
				action="?/saveToday"
use:enhance={createSubmitHandler(
	(loading) => todayLoading = loading,
	(saved) => {
		todaySaved = saved;
		if (saved) {
			// 保存成功時に初期コンテンツを更新
			initialTodayContent = todayContent;
			allowNavigation = true;
		}
	}
)}
				slot="form"
			>
				<input
					type="hidden"
					name="date"
					value={formatDateStr(data.today.date)}
				/>
				{#if data.today.entry}
					<input type="hidden" name="id" value={data.today.entry.id} />
				{/if}
				<input type="hidden" name="selectedEntities" value={JSON.stringify(todaySelectedEntities)} />
				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					diaryEntities={data.today.entry?.diaryEntities || []}
					bind:value={todayContent}
					bind:selectedEntities={todaySelectedEntities}
					on:save={handleSave}
				/>

				<!-- Character count display -->
				<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{$_("diary.characterCount", { values: { count: todayCharacterCount } })}
					{#if todayCharacterCount >= 1000}
						<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
							({$_("diary.autoSummaryEligible")})
						</span>
					{/if}
				</div>

				<div class="sticky bottom-4 flex justify-end hidden sm:flex mt-4 z-10">
					<SaveButton loading={todayLoading} saved={todaySaved} />
				</div>
			</form>
		</DiaryCard>
		</div>

		<div bind:this={yesterdayCard}>
		<DiaryCard
			title={$_("diary.yesterday")}
			entry={data.yesterday.entry}
			showForm={true}
			href={`/${formatDateStr(data.yesterday.date)}`}
		>
			<form
				bind:this={yesterdayFormElement}
				method="POST"
				action="?/saveYesterday"
use:enhance={createSubmitHandler(
	(loading) => yesterdayLoading = loading,
	(saved) => {
		yesterdaySaved = saved;
		if (saved) {
			// 保存成功時に初期コンテンツを更新
			initialYesterdayContent = yesterdayContent;
			allowNavigation = true;
		}
	}
)}
				slot="form"
			>
				<input
					type="hidden"
					name="date"
					value={formatDateStr(data.yesterday.date)}
				/>
				{#if data.yesterday.entry}
					<input type="hidden" name="id" value={data.yesterday.entry.id} />
				{/if}
				<input type="hidden" name="selectedEntities" value={JSON.stringify(yesterdaySelectedEntities)} />
				<FormField
					type="textarea"
					label=""
					id="yesterday-content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					diaryEntities={data.yesterday.entry?.diaryEntities || []}
					bind:value={yesterdayContent}
					bind:selectedEntities={yesterdaySelectedEntities}
					on:save={handleYesterdaySave}
				/>

				<!-- Character count display -->
				<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{$_("diary.characterCount", { values: { count: yesterdayCharacterCount } })}
					{#if yesterdayCharacterCount >= 1000}
						<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
							({$_("diary.autoSummaryEligible")})
						</span>
					{/if}
				</div>

				<div class="sticky bottom-4 flex justify-end hidden sm:flex mt-4 z-10">
					<SaveButton loading={yesterdayLoading} saved={yesterdaySaved} />
				</div>
			</form>
		</DiaryCard>
		</div>

		<div bind:this={dayBeforeYesterdayCard}>
		<DiaryCard
			title={$_("diary.dayBeforeYesterday")}
			entry={data.dayBeforeYesterday.entry}
			showForm={true}
			href={`/${formatDateStr(data.dayBeforeYesterday.date)}`}
		>
			<form
				bind:this={dayBeforeYesterdayFormElement}
				method="POST"
				action="?/saveDayBeforeYesterday"
use:enhance={createSubmitHandler(
	(loading) => dayBeforeLoading = loading,
	(saved) => {
		dayBeforeSaved = saved;
		if (saved) {
			// 保存成功時に初期コンテンツを更新
			initialDayBeforeYesterdayContent = dayBeforeYesterdayContent;
			allowNavigation = true;
		}
	}
)}
				slot="form"
			>
				<input
					type="hidden"
					name="date"
					value={formatDateStr(data.dayBeforeYesterday.date)}
				/>
				{#if data.dayBeforeYesterday.entry}
					<input
						type="hidden"
						name="id"
						value={data.dayBeforeYesterday.entry.id}
					/>
				{/if}
				<input type="hidden" name="selectedEntities" value={JSON.stringify(dayBeforeYesterdaySelectedEntities)} />
				<FormField
					type="textarea"
					label=""
					id="day-before-yesterday-content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					diaryEntities={data.dayBeforeYesterday.entry?.diaryEntities || []}
					bind:value={dayBeforeYesterdayContent}
					bind:selectedEntities={dayBeforeYesterdaySelectedEntities}
					on:save={handleDayBeforeYesterdaySave}
				/>

				<!-- Character count display -->
				<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{$_("diary.characterCount", { values: { count: dayBeforeYesterdayCharacterCount } })}
					{#if dayBeforeYesterdayCharacterCount >= 1000}
						<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
							({$_("diary.autoSummaryEligible")})
						</span>
					{/if}
				</div>

				<div class="sticky bottom-4 flex justify-end hidden sm:flex mt-4 z-10">
					<SaveButton loading={dayBeforeLoading} saved={dayBeforeSaved} />
				</div>
			</form>
		</DiaryCard>
		</div>
	</div>

	<!-- PWA Install Button -->
	<PWAInstallButton />

	<!-- Fixed Save Button for Mobile (shows only the active section) -->
	<div class="fixed bottom-20 left-0 right-0 p-4 sm:hidden z-10 pointer-events-none">
		<div class="max-w-4xl mx-auto flex justify-end pointer-events-auto">
			{#if activeSection === "today"}
				<Button type="button" variant={todaySaved ? "success" : "primary"} size="md" disabled={todayLoading || todaySaved} on:click={handleSave}>
					<div class="flex items-center justify-center min-h-[1.25rem]">
						{#if todayLoading}
							<svg class="animate-spin -mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 714 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
							</svg>
							<span class="ml-1">{$_("diary.saving")}</span>
						{:else if todaySaved}
							<svg class="-mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
								<path stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" d="m9 12 2 2 4-4"/>
							</svg>
							<span class="ml-1">{$_("diary.saved")}</span>
						{:else}
							<span>{$_("diary.saveTodayDiary")}</span>
						{/if}
					</div>
				</Button>
			{:else if activeSection === "yesterday"}
				<Button type="button" variant={yesterdaySaved ? "success" : "primary"} size="md" disabled={yesterdayLoading || yesterdaySaved} on:click={handleYesterdaySave}>
					<div class="flex items-center justify-center min-h-[1.25rem]">
						{#if yesterdayLoading}
							<svg class="animate-spin -mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 714 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
							</svg>
							<span class="ml-1">{$_("diary.saving")}</span>
						{:else if yesterdaySaved}
							<svg class="-mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
								<path stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" d="m9 12 2 2 4-4"/>
							</svg>
							<span class="ml-1">{$_("diary.saved")}</span>
						{:else}
							<span>{$_("diary.saveYesterdayDiary")}</span>
						{/if}
					</div>
				</Button>
			{:else}
				<Button type="button" variant={dayBeforeSaved ? "success" : "primary"} size="md" disabled={dayBeforeLoading || dayBeforeSaved} on:click={handleDayBeforeYesterdaySave}>
					<div class="flex items-center justify-center min-h-[1.25rem]">
						{#if dayBeforeLoading}
							<svg class="animate-spin -mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 714 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
							</svg>
							<span class="ml-1">{$_("diary.saving")}</span>
						{:else if dayBeforeSaved}
							<svg class="-mr-1 h-4 w-4" fill="none" viewBox="0 0 24 24">
								<path stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" d="m9 12 2 2 4-4"/>
							</svg>
							<span class="ml-1">{$_("diary.saved")}</span>
						{:else}
							<span>{$_("diary.saveDayBeforeYesterdayDiary")}</span>
						{/if}
					</div>
				</Button>
			{/if}
		</div>
	</div>

	<!-- Spacer for fixed button on mobile -->
	<div class="h-32 sm:hidden"></div>
</div>

