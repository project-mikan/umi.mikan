<script lang="ts">
  import { _ } from "svelte-i18n";
  import type { DiaryEntry } from "#lib/grpc/diary/diary_pb.js";
  import Button from "../atoms/Button.svelte";
  import Card from "../atoms/Card.svelte";
  import type { Snippet } from "svelte";

  let {
    title,
    entry = null,
    showForm = false,
    onView = null,
    href = null,
    editor,
    emptyMessage,
  }: {
    title: string;
    entry?: DiaryEntry | null;
    showForm?: boolean;
    onView?: ((entry: DiaryEntry) => void) | null;
    href?: string | null;
    editor?: Snippet;
    emptyMessage?: Snippet;
  } = $props();

  let highlightedContent = $derived(entry ? entry.content : "");
</script>

<Card>
	<div class="flex justify-between items-center">
		<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100">
			{#if href}
				<a
					{href}
					class="hover:text-blue-600 dark:hover:text-blue-400 transition-colors duration-200 underline decoration-transparent hover:decoration-current"
				>
					{title}
				</a>
			{:else}
				{title}
			{/if}
		</h2>
	</div>

	{#if showForm}
		{@render editor?.()}
	{:else if entry}
		<div
			class="text-gray-700 dark:text-gray-300 whitespace-pre-wrap auto-phrase-target"
		>
			{highlightedContent}
		</div>
		{#if onView}
			<div class="mt-4">
				<Button
					variant="primary"
					size="sm"
					onclick={() => onView && onView(entry)}
				>
					{$_("diary.viewDetail")}
				</Button>
			</div>
		{/if}
	{:else}
		<p class="text-gray-500 dark:text-gray-400">
			{@render emptyMessage?.()}
		</p>
	{/if}
</Card>

