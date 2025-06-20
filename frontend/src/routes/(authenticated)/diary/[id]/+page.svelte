<script lang="ts">
import { goto } from "$app/navigation";
import type { PageData } from "./$types";

export let data: PageData;

function formatDate(ymd: { year: number; month: number; day: number }): string {
	return `${ymd.year}年${ymd.month}月${ymd.day}日`;
}

function editEntry() {
	const date = data.entry.date;
	if (date) {
		const dateStr = `${date.year}-${String(date.month).padStart(2, "0")}-${String(date.day).padStart(2, "0")}`;
		goto(`/diary/edit/${dateStr}`);
	}
}

function goBack() {
	goto("/diary");
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">日記詳細</h1>
		<button
			on:click={goBack}
			class="text-gray-600 hover:text-gray-800 font-medium"
		>
			戻る
		</button>
	</div>

	<div class="bg-white shadow rounded-lg p-6">
		<div class="flex justify-between items-center mb-6">
			<h2 class="text-2xl font-semibold text-gray-900">
				{data.entry.date ? formatDate(data.entry.date) : '日付不明'}
			</h2>
			<button
				on:click={editEntry}
				class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
			>
				編集
			</button>
		</div>

		<div class="prose max-w-none">
			<div class="whitespace-pre-wrap text-gray-700 leading-relaxed">
				{data.entry.content}
			</div>
		</div>
	</div>
</div>