<script lang="ts">
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import { _ } from "svelte-i18n";
import "$lib/i18n";
import type { ActionData } from "./$types";

export let form: ActionData;

let content = "";
let date = new Date().toISOString().split("T")[0];

function cancel() {
	goto("/diary");
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">{$_('create.title')}</h1>
		<button
			on:click={cancel}
			class="text-gray-600 hover:text-gray-800 font-medium"
		>
			{$_('diary.back')}
		</button>
	</div>

	<div class="bg-white shadow rounded-lg p-6">
		<form method="POST" use:enhance>
			{#if form?.error}
				<div class="mb-4 p-4 bg-red-50 border border-red-200 rounded">
					<p class="text-red-600">{form.error}</p>
				</div>
			{/if}

			<div class="mb-6">
				<label for="date" class="block text-sm font-medium text-gray-700 mb-2">
					{$_('create.date')}
				</label>
				<input
					type="date"
					id="date"
					name="date"
					bind:value={date}
					required
					class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
				/>
			</div>

			<div class="mb-6">
				<label for="content" class="block text-sm font-medium text-gray-700 mb-2">
					{$_('create.content')}
				</label>
				<textarea
					id="content"
					name="content"
					bind:value={content}
					required
					rows="12"
					placeholder={$_('diary.placeholder')}
					class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 resize-none"
				></textarea>
			</div>

			<div class="flex justify-end space-x-4">
				<button
					type="button"
					on:click={cancel}
					class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
				>
					{$_('diary.cancel')}
				</button>
				<button
					type="submit"
					class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md font-medium"
				>
					{$_('diary.save')}
				</button>
			</div>
		</form>
	</div>
</div>