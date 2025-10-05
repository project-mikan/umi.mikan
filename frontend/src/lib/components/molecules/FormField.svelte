<script lang="ts">
import { createEventDispatcher } from "svelte";
import Input from "../atoms/Input.svelte";
import Label from "../atoms/Label.svelte";
import Textarea from "../atoms/Textarea.svelte";
import type { DiaryEntityOutput } from "$lib/grpc/diary/diary_pb";

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
export let diaryEntities: DiaryEntityOutput[] = [];

const dispatch = createEventDispatcher();
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
			{diaryEntities}
			bind:value
			on:save={() => dispatch('save')}
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