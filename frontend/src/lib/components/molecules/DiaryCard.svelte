<script lang="ts">
import { _ } from "svelte-i18n";
import type { DiaryEntry } from "$lib/grpc";
import Button from "../atoms/Button.svelte";
import Card from "../atoms/Card.svelte";

export let title: string;
export let entry: DiaryEntry | null = null;
export let showForm = false;
export let onView: ((entry: DiaryEntry) => void) | null = null;

function formatContentWithLineBreaks(content: string): string {
	return content.replace(/\n/g, "<br>");
}
</script>

<Card>
	<div class="flex justify-between items-center mb-4">
		<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100">
			{title}
		</h2>
	</div>

	{#if showForm}
		<slot name="form" />
	{:else if entry}
		<div class="text-gray-700 dark:text-gray-300">{@html formatContentWithLineBreaks(entry.content || '')}</div>
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
		<p class="text-gray-500 dark:text-gray-400">
			<slot name="empty-message" />
		</p>
	{/if}
</Card>