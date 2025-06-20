<script lang="ts">
import { goto } from "$app/navigation";
import type { ActionData } from "./$types.ts";

export let form: ActionData;

const content = "";
const date = new Date().toISOString().split("T")[0];

function cancel() {
	goto("/diary");
}
</script>

<div class="max-w-4xl mx-auto">
	<div class="bg-white shadow rounded-lg">
		<div class="px-6 py-4 border-b border-gray-200">
			<h1 class="text-xl font-semibold text-gray-900">新しい日記を作成</h1>
		</div>

		{#if form?.error}
			<div class="mx-6 mt-4 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
				{form.error}
			</div>
		{/if}

		<form method="POST" class="p-6 space-y-6" use:enhance>
			<div>
				<label for="date" class="block text-sm font-medium text-gray-700 mb-2">
					日付
				</label>
				<input
					type="date"
					id="date"
					name="date"
					bind:value={date}
					required
					class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
				/>
			</div>

			<div>
				<label for="content" class="block text-sm font-medium text-gray-700 mb-2">
					内容
				</label>
				<textarea
					id="content"
					name="content"
					bind:value={content}
					rows="15"
					required
					class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
					placeholder="今日の出来事を書いてください..."
				></textarea>
			</div>

			<div class="flex justify-end space-x-3">
				<button
					type="button"
					on:click={cancel}
					class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
				>
					キャンセル
				</button>
				<button
					type="submit"
					class="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
				>
					保存
				</button>
			</div>
		</form>
	</div>
</div>