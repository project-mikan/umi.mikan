<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";
  import Input from "../atoms/Input.svelte";
  import Label from "../atoms/Label.svelte";
  import Textarea from "../atoms/Textarea.svelte";
  import type { DiaryHighlight } from "#lib/types/highlight.js";

  interface Props {
    type?: "input" | "textarea";
    inputType?: "text" | "email" | "password" | "date";
    label: string;
    id: string;
    name: string;
    value?: string;
    placeholder?: string;
    required?: boolean;
    disabled?: boolean;
    autocomplete?: HTMLInputAttributes["autocomplete"];
    rows?: number;
    srOnlyLabel?: boolean;
    diaryHighlights?: DiaryHighlight[];
    searchKeyword?: string;
    // textarea の保存・自動保存のタイミングで呼ぶ
    onsave?: () => void;
    onautosave?: () => void;
  }

  let {
    type = "input",
    inputType = "text",
    label,
    id,
    name,
    value = $bindable(""),
    placeholder = "",
    required = false,
    disabled = false,
    autocomplete = "",
    rows = 4,
    srOnlyLabel = false,
    diaryHighlights = [],
    searchKeyword = "",
    onsave = () => {},
    onautosave = () => {},
  }: Props = $props();
</script>

<div class="mb-4">
	<Label htmlFor={id} {required} srOnly={srOnlyLabel}>
		{label}
	</Label>

	{#if type === 'textarea'}
		<Textarea
			{id}
			{name}
			{placeholder}
			{required}
			{disabled}
			{rows}
			{diaryHighlights}
			{searchKeyword}
			bind:value
			{onsave}
			{onautosave}
		/>
	{:else}
		<Input
			type={inputType}
			{id}
			{name}
			{placeholder}
			{required}
			{disabled}
			{autocomplete}
			bind:value
		/>
	{/if}
</div>
