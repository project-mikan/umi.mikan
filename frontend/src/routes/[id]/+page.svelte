<script lang="ts">
import { _, locale } from "svelte-i18n";
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import "$lib/i18n";
import Button from "$lib/components/atoms/Button.svelte";
import SaveButton from "$lib/components/atoms/SaveButton.svelte";
import DiaryCard from "$lib/components/molecules/DiaryCard.svelte";
import DiaryNavigation from "$lib/components/molecules/DiaryNavigation.svelte";
import FormField from "$lib/components/molecules/FormField.svelte";
import Modal from "$lib/components/molecules/Modal.svelte";
import PastEntriesLinks from "$lib/components/molecules/PastEntriesLinks.svelte";
import { getDayOfWeekKey } from "$lib/utils/date-utils";
import { createSubmitHandler } from "$lib/utils/form-utils";
import type { ActionData, PageData } from "./$types";

export let data: PageData;
export let form: ActionData;

$: content = data.entry?.content || "";
let formElement: HTMLFormElement;
let _showDeleteConfirm = false;
let loading = false;
let saved = false;
let summaryGenerating = false;
let summary: {
	id: string;
	diaryId: string;
	date: { year: number; month: number; day: number };
	summary: string;
	createdAt: number;
} | null = null;
let showSummary = false;

// Check if user has LLM key configured
$: existingLLMKey = data.user?.llmKeys?.find((key) => key.llmProvider === 1);
$: hasLLMKey = !!existingLLMKey;
$: autoSummaryDisabled = !existingLLMKey?.autoSummaryDaily;

// Check if the diary date is not today (only allow summary generation for past entries)
$: isNotToday = (() => {
	const today = new Date();
	const diaryDate = new Date(
		data.date.year,
		data.date.month - 1,
		data.date.day,
	);
	return diaryDate < today;
})();

// Character count calculation
$: characterCount = content ? content.length : 0;

// Reactive date formatting function
$: _formatDate = (ymd: {
	year: number;
	month: number;
	day: number;
}): string => {
	const dayOfWeekKey = getDayOfWeekKey(ymd);
	const dayOfWeek = $_(`date.dayOfWeek.${dayOfWeekKey}`);
	return $_("date.format.yearMonthDayWithDayOfWeek", {
		values: {
			year: ymd.year,
			month: ymd.month,
			day: ymd.day,
			dayOfWeek: dayOfWeek,
		},
	});
};

function _formatDateStr(ymd: {
	year: number;
	month: number;
	day: number;
}): string {
	return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
}

function _goBack() {
	goto("/");
}

function _goToMonthly() {
	const year = data.date.year;
	const month = String(data.date.month).padStart(2, "0");
	goto(`/monthly/${year}/${month}`);
}

function _handleSave() {
	formElement?.requestSubmit();
}

function _confirmDelete() {
	_showDeleteConfirm = true;
}

function _cancelDelete() {
	_showDeleteConfirm = false;
}

function _handleDelete() {
	const form = document.createElement("form");
	form.method = "POST";
	form.action = "?/delete";
	document.body.appendChild(form);
	form.submit();
}

async function _generateSummary() {
	if (!data.entry?.content || summaryGenerating) return;

	summaryGenerating = true;
	try {
		const response = await fetch("/api/diary/summary/generate-daily", {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify({
				diaryId: data.entry.id,
				content: data.entry.content,
				date: data.date,
			}),
		});

		if (!response.ok) {
			const errorData = await response.json().catch(() => ({}));
			throw new Error(errorData.message || "要約の生成に失敗しました");
		}

		const result = await response.json();
		summary = result;
		showSummary = true;
	} catch (error) {
		console.error("Summary generation failed:", error);
		alert(error instanceof Error ? error.message : "要約の生成に失敗しました");
	} finally {
		summaryGenerating = false;
	}
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">{$_("diary.title")}</h1>
		<div class="flex gap-2">
			{#if summary}
				<button
					on:click={() => showSummary = !showSummary}
					class="px-4 py-2 {showSummary ? 'bg-gray-600 hover:bg-gray-700' : 'bg-blue-600 hover:bg-blue-700'} text-white rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
				>
					{showSummary ? $_("diary.summary.hide") : $_("diary.summary.view")}
				</button>
			{/if}
			{#if data.entry && hasLLMKey && autoSummaryDisabled && characterCount >= 1000 && isNotToday}
				<button
					on:click={_generateSummary}
					disabled={summaryGenerating}
					class="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2"
				>
					{summaryGenerating ? $_("diary.generatingSummary") : $_("diary.generateSummary")}
				</button>
			{/if}
			<button
				on:click={_goToMonthly}
				class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium"
			>
				{$_("diary.viewThisMonth")}
			</button>
		</div>
	</div>

	<div class="space-y-6">
		<DiaryNavigation currentDate={data.date} />

		<!-- Summary display area -->
		{#if showSummary && summary}
			<div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700">
				<div class="p-6">
					<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
						{$_("diary.summary.title")} - {_formatDate(data.date)}
					</h2>
					<div class="prose dark:prose-invert max-w-none">
						<p class="text-gray-700 dark:text-gray-300 whitespace-pre-wrap leading-relaxed">
							{summary.summary}
						</p>
					</div>
					<div class="mt-6 flex justify-between items-center text-sm text-gray-500 dark:text-gray-400">
						<span>
							{$_("common.createdAt")}: {new Date(summary.createdAt).toLocaleString()}
						</span>
					</div>
				</div>
			</div>
		{/if}

		<DiaryCard
			title={_formatDate(data.date)}
			entry={data.entry}
			showForm={true}
		>
			<form
				bind:this={formElement}
				method="POST"
				action="?/save"
use:enhance={createSubmitHandler((l) => loading = l, (s) => saved = s)}
				slot="form"
			>
				<input type="hidden" name="date" value={_formatDateStr(data.date)} />
				{#if data.entry}
					<input type="hidden" name="id" value={data.entry.id} />
				{/if}
				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					bind:value={content}
					on:save={_handleSave}
				/>
				{#if form?.error}
					<div class="mt-2 text-sm text-red-600 dark:text-red-400">
						{form.error}
					</div>
				{/if}

				<!-- Character count display -->
				<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{$_("diary.characterCount", { values: { count: characterCount } })}
					{#if characterCount >= 1000}
						<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
							({$_("diary.autoSummaryEligible")})
						</span>
					{/if}
				</div>


				<div class="flex justify-between">
					<div>
						{#if data.entry}
							<Button
								type="button"
								variant="danger"
								size="md"
								on:click={_confirmDelete}
							>
								{$_("diary.delete")}
							</Button>
						{/if}
					</div>
					<SaveButton {loading} {saved} />
				</div>
			</form>
		</DiaryCard>

		<PastEntriesLinks pastEntries={data.pastEntries} />
	</div>
</div>

<Modal
	isOpen={_showDeleteConfirm}
	title={$_("edit.deleteConfirm")}
	confirmText={$_("diary.delete")}
	cancelText={$_("diary.cancel")}
	variant="danger"
	onConfirm={_handleDelete}
	onCancel={_cancelDelete}
>
	<p class="text-sm text-gray-500 dark:text-gray-400">
		{$_("edit.deleteMessage")}
	</p>
</Modal>

