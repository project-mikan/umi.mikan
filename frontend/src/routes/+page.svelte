<script lang="ts">
import { goto } from "$app/navigation";
import "$lib/i18n";
import type { DiaryEntry, YMD } from "$lib/grpc";
import type { PageData } from "./$types";

export let data: PageData;

let _todayContent = data.today.entry?.content || "";
let _yesterdayContent = data.yesterday.entry?.content || "";
let _dayBeforeYesterdayContent = data.dayBeforeYesterday.entry?.content || "";
let formElement: HTMLFormElement;
let yesterdayFormElement: HTMLFormElement;
let dayBeforeYesterdayFormElement: HTMLFormElement;

function _getMonthlyUrl(): string {
	const now = new Date();
	return `/monthly/${now.getFullYear()}/${now.getMonth() + 1}`;
}

function formatDateStr(ymd: YMD): string {
	return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
}

function _viewEntry(entry: DiaryEntry) {
	const date = entry.date;
	if (date) {
		const dateStr = formatDateStr(date);
		goto(`/${dateStr}`);
	}
}

function _handleSave() {
	formElement?.requestSubmit();
}

function _handleYesterdaySave() {
	yesterdayFormElement?.requestSubmit();
}

function _handleDayBeforeYesterdaySave() {
	dayBeforeYesterdayFormElement?.requestSubmit();
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">{$_('diary.title')}</h1>
		<div class="flex gap-3">
			<Button variant="success" size="md">
				<a href={_getMonthlyUrl()} class="text-white">
					{$_('diary.thisMonth')}
				</a>
			</Button>
			<Button variant="gray" size="md">
				<a href="/search" class="text-white">
					{$_('diary.search')}
				</a>
			</Button>
		</div>
	</div>

	<div class="space-y-6">
		<DiaryCard
			title={$_('diary.today')}
			entry={data.today.entry}
			showForm={true}
		>
			<form bind:this={formElement} method="POST" action="?/saveToday" use:enhance={(() => {
				return async ({ result }) => {
					if (result.type === 'success') {
						window.location.reload();
					}
				};
			})} slot="form">
				<input type="hidden" name="date" value={formatDateStr(data.today.date)} />
				{#if data.today.entry}
					<input type="hidden" name="id" value={data.today.entry.id} />
				{/if}
				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_('diary.placeholder')}
					rows={8}
					bind:value={_todayContent}
					on:save={_handleSave}
				/>
				<div class="flex justify-end">
					<Button type="submit" variant="primary" size="md">
						{$_('diary.save')}
					</Button>
				</div>
			</form>
		</DiaryCard>

		<DiaryCard
			title={$_('diary.yesterday')}
			entry={data.yesterday.entry}
			showForm={true}
		>
			<form bind:this={yesterdayFormElement} method="POST" action="?/saveYesterday" use:enhance={(() => {
				return async ({ result }) => {
					if (result.type === 'success') {
						window.location.reload();
					}
				};
			})} slot="form">
				<input type="hidden" name="date" value={formatDateStr(data.yesterday.date)} />
				{#if data.yesterday.entry}
					<input type="hidden" name="id" value={data.yesterday.entry.id} />
				{/if}
				<FormField
					type="textarea"
					label=""
					id="yesterday-content"
					name="content"
					placeholder={$_('diary.placeholder')}
					rows={8}
					bind:value={_yesterdayContent}
					on:save={_handleYesterdaySave}
				/>
				<div class="flex justify-end">
					<Button type="submit" variant="primary" size="md">
						{$_('diary.save')}
					</Button>
				</div>
			</form>
		</DiaryCard>

		<DiaryCard
			title={$_('diary.dayBeforeYesterday')}
			entry={data.dayBeforeYesterday.entry}
			showForm={true}
		>
			<form bind:this={dayBeforeYesterdayFormElement} method="POST" action="?/saveDayBeforeYesterday" use:enhance={(() => {
				return async ({ result }) => {
					if (result.type === 'success') {
						window.location.reload();
					}
				};
			})} slot="form">
				<input type="hidden" name="date" value={formatDateStr(data.dayBeforeYesterday.date)} />
				{#if data.dayBeforeYesterday.entry}
					<input type="hidden" name="id" value={data.dayBeforeYesterday.entry.id} />
				{/if}
				<FormField
					type="textarea"
					label=""
					id="day-before-yesterday-content"
					name="content"
					placeholder={$_('diary.placeholder')}
					rows={8}
					bind:value={_dayBeforeYesterdayContent}
					on:save={_handleDayBeforeYesterdaySave}
				/>
				<div class="flex justify-end">
					<Button type="submit" variant="primary" size="md">
						{$_('diary.save')}
					</Button>
				</div>
			</form>
		</DiaryCard>
	</div>
</div>