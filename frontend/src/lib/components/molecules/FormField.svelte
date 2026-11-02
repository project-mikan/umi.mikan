<script lang="ts">
  import Input from "../atoms/Input.svelte";
  import Label from "../atoms/Label.svelte";
  import Textarea from "../atoms/Textarea.svelte";
  import type { DiaryHighlight } from "#lib/types/highlight.js";

  export let type: "input" | "textarea" = "input";
  export let inputType: "text" | "email" | "password" | "date" = "text";
  export let label: string;
  export let id: string;
  export let name: string;
  export let value = "";
  export let placeholder = "";
  export let required = false;
  export let disabled = false;
  export let autocomplete = "";
  export let rows = 4;
  export let srOnlyLabel = false;
  export let diaryHighlights: DiaryHighlight[] = [];
  export let searchKeyword = "";
  // textarea の保存・自動保存のタイミングで呼ぶ
  export let onsave: () => void = () => {};
  export let onautosave: () => void = () => {};
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
