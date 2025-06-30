<script lang="ts">
import { enhance } from "$app/forms";
import { _ } from "svelte-i18n";
import "$lib/i18n";
import DiaryForm from "$lib/components/molecules/DiaryForm.svelte";
import Modal from "$lib/components/molecules/Modal.svelte";
import type { ActionData, PageData } from "./$types";

export let data: PageData;
export let form: ActionData;

let content = data.entry.content;
let date = data.entry.date
	? `${data.entry.date.year}-${String(data.entry.date.month).padStart(2, "0")}-${String(data.entry.date.day).padStart(2, "0")}`
	: new Date().toISOString().split("T")[0];
let showDeleteConfirm = false;

function confirmDelete() {
	showDeleteConfirm = true;
}

function cancelDelete() {
	showDeleteConfirm = false;
}

function handleDelete() {
	const form = document.createElement("form");
	form.method = "POST";
	form.action = "?/delete";
	document.body.appendChild(form);
	form.submit();
}
</script>

<form method="POST" action="?/update" use:enhance>
	<DiaryForm
		title={$_('edit.title')}
		bind:content
		bind:date
		error={form?.error}
		showDeleteButton={true}
		onDelete={confirmDelete}
	/>
</form>

<Modal
	isOpen={showDeleteConfirm}
	title={$_('edit.deleteConfirm')}
	confirmText={$_('diary.delete')}
	cancelText={$_('diary.cancel')}
	variant="danger"
	onConfirm={handleDelete}
	onCancel={cancelDelete}
>
	<p class="text-sm text-gray-500">
		{$_('edit.deleteMessage')}
	</p>
</Modal>