<script lang="ts">
import { _ } from "svelte-i18n";
import { enhance } from "$app/forms";
import "$lib/i18n";
import type { ActionData } from "./$types";

export let form: ActionData;

let name = form?.name || "";
let memo = form?.memo || "";
let isSubmitting = false;
</script>

<svelte:head>
	<title>{$_("entity.create.title")} - umi.mikan</title>
</svelte:head>

<div class="max-w-2xl mx-auto p-6">
	<div class="mb-6">
		<a
			href="/entities"
			class="text-blue-600 dark:text-blue-400 hover:underline flex items-center gap-1"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M15 19l-7-7 7-7"
				></path>
			</svg>
			{$_("entity.list.title")}
		</a>
	</div>

	<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-6">
		{$_("entity.create.title")}
	</h1>

	<!-- エラーメッセージ -->
	{#if form?.error}
		<div
			class="bg-red-100 dark:bg-red-900/20 border border-red-400 dark:border-red-600 text-red-700 dark:text-red-400 px-4 py-3 rounded mb-6"
		>
			<p>{$_(form.error)}</p>
			{#if "errorDetail" in form && form.errorDetail}
				<p class="mt-2 text-sm font-mono">{form.errorDetail}</p>
			{/if}
		</div>
	{/if}

	<form
		method="POST"
		action="?/create"
		use:enhance={() => {
			isSubmitting = true;
			return async ({ update }) => {
				await update();
				isSubmitting = false;
			};
		}}
		class="space-y-6"
	>
		<!-- 名前 -->
		<div>
			<label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
				{$_("entity.create.name")}
			</label>
			<input
				type="text"
				id="name"
				name="name"
				bind:value={name}
				required
				class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
				placeholder={$_("entity.create.namePlaceholder")}
			/>
		</div>

		<!-- メモ -->
		<div>
			<label for="memo" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
				{$_("entity.create.memo")}
			</label>
			<textarea
				id="memo"
				name="memo"
				bind:value={memo}
				rows="4"
				class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 auto-phrase-target"
				placeholder={$_("entity.create.memoPlaceholder")}
			></textarea>
		</div>

		<!-- ボタン -->
		<div class="flex gap-4">
			<button
				type="submit"
				disabled={isSubmitting}
				class="flex-1 px-6 py-3 bg-blue-600 dark:bg-blue-500 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
			>
				{isSubmitting ? $_("common.loading") : $_("entity.create.submit")}
			</button>
			<a
				href="/entities"
				class="flex-1 px-6 py-3 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 text-center"
			>
				{$_("common.cancel")}
			</a>
		</div>
	</form>
</div>
