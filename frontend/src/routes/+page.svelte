<script lang="ts">
import { _ } from "svelte-i18n";
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import "$lib/i18n";
import Button from "$lib/components/atoms/Button.svelte";
import DiaryCard from "$lib/components/molecules/DiaryCard.svelte";
import FormField from "$lib/components/molecules/FormField.svelte";
import TimeProgressBar from "$lib/components/molecules/TimeProgressBar.svelte";
import type { DiaryEntry, YMD } from "$lib/grpc";
import type { PageData } from "./$types";

export let data: PageData;

let todayContent = data.today.entry?.content || "";
let yesterdayContent = data.yesterday.entry?.content || "";
let dayBeforeYesterdayContent = data.dayBeforeYesterday.entry?.content || "";
let formElement: HTMLFormElement;
let yesterdayFormElement: HTMLFormElement;
let dayBeforeYesterdayFormElement: HTMLFormElement;

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
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">{$_("diary.title")}</h1>
	</div>

	<div class="mb-8">
		<TimeProgressBar />
	</div>

	<div class="space-y-6">
		<DiaryCard
			title={$_("diary.today")}
			entry={data.today.entry}
			showForm={true}
		>
			<form
				bind:this={formElement}
				method="POST"
				action="?/saveToday"
				use:enhance={() => {
					return async ({ result }) => {
						if (result.type === "success") {
							window.location.reload();
						}
					};
				}}
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
				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					bind:value={todayContent}
					on:save={handleSave}
				/>
				<div class="flex justify-end">
					<Button type="submit" variant="primary" size="md">
						{$_("diary.save")}
					</Button>
				</div>
			</form>
		</DiaryCard>

		<DiaryCard
			title={$_("diary.yesterday")}
			entry={data.yesterday.entry}
			showForm={true}
		>
			<form
				bind:this={yesterdayFormElement}
				method="POST"
				action="?/saveYesterday"
				use:enhance={() => {
					return async ({ result }) => {
						if (result.type === "success") {
							window.location.reload();
						}
					};
				}}
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
				<FormField
					type="textarea"
					label=""
					id="yesterday-content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					bind:value={yesterdayContent}
					on:save={handleYesterdaySave}
				/>
				<div class="flex justify-end">
					<Button type="submit" variant="primary" size="md">
						{$_("diary.save")}
					</Button>
				</div>
			</form>
		</DiaryCard>

		<DiaryCard
			title={$_("diary.dayBeforeYesterday")}
			entry={data.dayBeforeYesterday.entry}
			showForm={true}
		>
			<form
				bind:this={dayBeforeYesterdayFormElement}
				method="POST"
				action="?/saveDayBeforeYesterday"
				use:enhance={() => {
					return async ({ result }) => {
						if (result.type === "success") {
							window.location.reload();
						}
					};
				}}
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
				<FormField
					type="textarea"
					label=""
					id="day-before-yesterday-content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					bind:value={dayBeforeYesterdayContent}
					on:save={handleDayBeforeYesterdaySave}
				/>
				<div class="flex justify-end">
					<Button type="submit" variant="primary" size="md">
						{$_("diary.save")}
					</Button>
				</div>
			</form>
		</DiaryCard>
	</div>
</div>

