<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import DiaryForm from "$lib/components/molecules/DiaryForm.svelte";
import type { ActionData } from "./$types";

export let form: ActionData;

let content = "";
let date = new Date().toISOString().split("T")[0];
</script>

<DiaryForm
	title={$_('create.title')}
	bind:content
	bind:date
	error={form?.error}
	showDeleteButton={false}
/>