<script lang="ts">
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
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
		<h1 class="text-3xl font-bold text-gray-900">新しい日記を書く</h1>
		<button
			on:click={cancel}
			class="text-gray-600 hover:text-gray-800 font-medium"
		>
			戻る
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
					日付
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
					内容
				</label>
				<textarea
					id="content"
					name="content"
					bind:value={content}
					required
					rows="12"
					placeholder="今日の出来事を書いてください..."
					class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 resize-none"
				></textarea>
			</div>

			<div class="flex justify-end space-x-4">
				<button
					type="button"
					on:click={cancel}
					class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50"
				>
					キャンセル
				</button>
				<button
					type="submit"
					class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md font-medium"
				>
					保存
				</button>
			</div>
		</form>
	</div>
</div>