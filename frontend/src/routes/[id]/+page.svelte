<script lang="ts">
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import { _ } from "svelte-i18n";
import "$lib/i18n";
import Button from "$lib/components/atoms/Button.svelte";
import DiaryCard from "$lib/components/molecules/DiaryCard.svelte";
import FormField from "$lib/components/molecules/FormField.svelte";
import type { ActionData, PageData } from "./$types";

export let data: PageData;
export let form: ActionData;

let content = data.entry?.content || "";
let formElement: HTMLFormElement;

function formatDate(ymd: { year: number; month: number; day: number }): string {
	return `${ymd.year}年${ymd.month}月${ymd.day}日`;
}

function formatDateStr(ymd: {
	year: number;
	month: number;
	day: number;
}): string {
	return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
}

function goBack() {
	goto("/");
}

function handleSave() {
	formElement?.requestSubmit();
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">{$_('diary.title')}</h1>
		<button
			on:click={goBack}
			class="text-gray-600 hover:text-gray-800 font-medium"
		>
			{$_('diary.back')}
		</button>
	</div>

	<div class="space-y-6">
		<DiaryCard
			title={formatDate(data.date)}
			date={data.date}
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
				<input type="hidden" name="date" value={formatDateStr(data.date)} />
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
					on:save={handleSave}
				/>
				{#if form?.error}
					<div class="mt-2 text-sm text-red-600">
						{form.error}
					</div>
				{/if}
				<div class="flex justify-end">
					<Button type="submit" variant="primary" size="md">
						{$_('diary.save')}
					</Button>
				</div>
			</form>
		</DiaryCard>
	</div>
</div>