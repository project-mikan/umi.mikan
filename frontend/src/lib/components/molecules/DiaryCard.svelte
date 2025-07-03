<script lang="ts">
import type { DiaryEntry, YMD } from "$lib/grpc";
import { _ } from "svelte-i18n";
import Button from "../atoms/Button.svelte";
import Card from "../atoms/Card.svelte";

export let title: string;
export let date: YMD;
export let entry: DiaryEntry | null = null;
export let showForm = false;
export const content = "";
export const isEditable = false;
export let onView: ((entry: DiaryEntry) => void) | null = null;

function formatDate(ymd: YMD): string {
	return `${ymd.year}年${ymd.month}月${ymd.day}日`;
}

function setText(node: HTMLElement, text: string) {
	node.textContent = text;
	return {
		update(newText: string) {
			node.textContent = newText;
		},
	};
}
</script>

<Card>
	<div class="flex justify-between items-center mb-4">
		<h2 class="text-xl font-semibold text-gray-900">
			{title} ({formatDate(date)})
		</h2>
	</div>

	{#if showForm}
		<slot name="form" />
	{:else if entry}
		<div class="text-gray-700 whitespace-pre-wrap" use:setText={entry.content || ''}></div>
		{#if onView}
			<div class="mt-4">
				<Button
					variant="primary"
					size="sm"
					on:click={() => onView && onView(entry)}
				>
					{$_('diary.viewDetail')}
				</Button>
			</div>
		{/if}
	{:else}
		<p class="text-gray-500">
			<slot name="empty-message" />
		</p>
	{/if}
</Card>