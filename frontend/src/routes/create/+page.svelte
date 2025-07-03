<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import DiaryForm from "$lib/components/molecules/DiaryForm.svelte";
import type { ActionData, PageData } from "./$types";

export let form: ActionData;
export let data: PageData;

let content = "";
let date = data.defaultDate;
</script>

<DiaryForm
	title={$_('create.title')}
	bind:content
	bind:date
	error={form?.error}
	showDeleteButton={false}
/>