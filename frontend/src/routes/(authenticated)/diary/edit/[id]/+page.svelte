<script lang="ts">
import { goto } from "$app/navigation";
import type { ActionData, PageData } from "./$types.ts";

export let data: PageData;
export let form: ActionData;

const content = data.entry.content;
const date = data.entry.date
	? `${data.entry.date.year}-${String(data.entry.date.month).padStart(2, "0")}-${String(data.entry.date.day).padStart(2, "0")}`
	: new Date().toISOString().split("T")[0];
let showDeleteConfirm = false;

function cancel() {
	goto("/diary");
}

function confirmDelete() {
	showDeleteConfirm = true;
}

function cancelDelete() {
	showDeleteConfirm = false;
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">日記を編集</h1>
		<button
			on:click={cancel}
			class="text-gray-600 hover:text-gray-800 font-medium"
		>
			戻る
		</button>
	</div>

	<div class="bg-white shadow rounded-lg p-6">
		<form method="POST" action="?/update" use:enhance>
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
					class="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 resize-none"
				></textarea>
			</div>

			<div class="flex justify-between">
				<button
					type="button"
					on:click={confirmDelete}
					class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-md font-medium"
				>
					削除
				</button>

				<div class="flex space-x-4">
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
						更新
					</button>
				</div>
			</div>
		</form>
	</div>
</div>

<!-- 削除確認モーダル -->
{#if showDeleteConfirm}
	<div class="fixed inset-0 z-50 overflow-y-auto">
		<div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
			<div class="fixed inset-0 transition-opacity" aria-hidden="true">
				<div class="absolute inset-0 bg-gray-500 opacity-75"></div>
			</div>

			<span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

			<div class="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
				<div class="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
					<div class="sm:flex sm:items-start">
						<div class="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10">
							<svg class="h-6 w-6 text-red-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
							</svg>
						</div>
						<div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
							<h3 class="text-lg leading-6 font-medium text-gray-900">
								日記を削除
							</h3>
							<div class="mt-2">
								<p class="text-sm text-gray-500">
									この日記を削除してもよろしいですか？この操作は取り消せません。
								</p>
							</div>
						</div>
					</div>
				</div>
				<div class="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
					<form method="POST" action="?/delete" use:enhance>
						<button
							type="submit"
							class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-red-600 text-base font-medium text-white hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 sm:ml-3 sm:w-auto sm:text-sm"
						>
							削除
						</button>
					</form>
					<button
						type="button"
						on:click={cancelDelete}
						class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm"
					>
						キャンセル
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}