<script lang="ts">
import { _ } from "svelte-i18n";
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import "$lib/i18n";
import Button from "$lib/components/atoms/Button.svelte";
import DiaryCard from "$lib/components/molecules/DiaryCard.svelte";
import DiaryNavigation from "$lib/components/molecules/DiaryNavigation.svelte";
import FormField from "$lib/components/molecules/FormField.svelte";
import Modal from "$lib/components/molecules/Modal.svelte";
import PastEntriesLinks from "$lib/components/molecules/PastEntriesLinks.svelte";
import { getDayOfWeekKey } from "$lib/utils/date-utils";
import type { ActionData, PageData } from "./$types";

export let data: PageData;
export let form: ActionData;

let content = data.entry?.content || "";
let formElement: HTMLFormElement;
let _showDeleteConfirm = false;

function _formatDate(ymd: {
	year: number;
	month: number;
	day: number;
}): string {
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
}

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
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">{$_('diary.title')}</h1>
		<div class="flex gap-4">
			<button
				on:click={_goToMonthly}
				class="text-blue-600 hover:text-blue-800 font-medium"
			>
				{$_('diary.viewThisMonth')}
			</button>
			<button
				on:click={_goBack}
				class="text-gray-600 hover:text-gray-800 font-medium"
			>
				{$_('diary.back')}
			</button>
		</div>
	</div>

	<div class="space-y-6">
		<DiaryNavigation currentDate={data.date} />
		<DiaryCard
			title={_formatDate(data.date)}
			entry={data.entry}
			showForm={true}
		>
			<form bind:this={formElement} method="POST" action="?/save" use:enhance={(() => {
				return async ({ result }) => {
					if (result.type === 'success') {
						window.location.reload();
					}
				};
			})} slot="form">
				<input type="hidden" name="date" value={_formatDateStr(data.date)} />
				{#if data.entry}
					<input type="hidden" name="id" value={data.entry.id} />
				{/if}
				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_('diary.placeholder')}
					rows={8}
					bind:value={content}
					on:save={_handleSave}
				/>
				{#if form?.error}
					<div class="mt-2 text-sm text-red-600">
						{form.error}
					</div>
				{/if}
				<div class="flex justify-between">
					<div>
						{#if data.entry}
							<Button type="button" variant="danger" size="md" on:click={_confirmDelete}>
								{$_('diary.delete')}
							</Button>
						{/if}
					</div>
					<Button type="submit" variant="primary" size="md">
						{$_('diary.save')}
					</Button>
				</div>
			</form>
		</DiaryCard>
		
		<PastEntriesLinks pastEntries={data.pastEntries} />
	</div>
</div>

<Modal
	isOpen={_showDeleteConfirm}
	title={$_('edit.deleteConfirm')}
	confirmText={$_('diary.delete')}
	cancelText={$_('diary.cancel')}
	variant="danger"
	onConfirm={_handleDelete}
	onCancel={_cancelDelete}
>
	<p class="text-sm text-gray-500">
		{$_('edit.deleteMessage')}
	</p>
</Modal>