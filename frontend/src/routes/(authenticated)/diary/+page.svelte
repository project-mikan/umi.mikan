<script lang="ts">
import { goto } from "$app/navigation";
import type { PageData } from "./$types";
import type { DiaryEntry } from "$lib/grpc";

export let data: PageData;

function formatDate(ymd: { year: number; month: number; day: number }): string {
	return `${ymd.year}年${ymd.month}月${ymd.day}日`;
}

function createEntry() {
	goto("/diary/create");
}

function editEntry(entry: DiaryEntry) {
	const date = entry.date;
	if (date) {
		const dateStr = `${date.year}-${String(date.month).padStart(2, '0')}-${String(date.day).padStart(2, '0')}`;
		goto(`/diary/edit/${dateStr}`);
	}
}

function viewEntry(entry: DiaryEntry) {
	const date = entry.date;
	if (date) {
		const dateStr = `${date.year}-${String(date.month).padStart(2, '0')}-${String(date.day).padStart(2, '0')}`;
		goto(`/diary/${dateStr}`);
	}
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900">日記一覧</h1>
		<div class="flex gap-3">
			<a
				href="/diary/search"
				class="bg-gray-600 hover:bg-gray-700 text-white font-bold py-2 px-4 rounded"
			>
				検索
			</a>
			<button
				on:click={createEntry}
				class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
			>
				新しい日記を書く
			</button>
		</div>
	</div>

	{#if !data.entries.entries || data.entries.entries.length === 0}
		<div class="bg-white shadow rounded-lg p-6 text-center">
			<p class="text-gray-600">まだ日記がありません。</p>
			<button
				on:click={createEntry}
				class="mt-4 bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
			>
				最初の日記を書く
			</button>
		</div>
	{:else}
		<div class="grid gap-6">
			{#each ((data.entries.entries || []) as DiaryEntry[]) as entry (entry.id)}
				<div class="bg-white shadow rounded-lg p-6 hover:shadow-lg transition-shadow">
					<div class="flex justify-between items-start mb-4">
						<h2 class="text-xl font-semibold text-gray-900">
							{entry.date ? formatDate(entry.date) : '日付不明'}
						</h2>
						<div class="flex space-x-2">
							<button
								on:click={() => viewEntry(entry)}
								class="text-blue-600 hover:text-blue-800 font-medium"
							>
								詳細
							</button>
							<button
								on:click={() => editEntry(entry)}
								class="text-green-600 hover:text-green-800 font-medium"
							>
								編集
							</button>
						</div>
					</div>
					<div class="text-gray-700">
						<p class="line-clamp-3">
							{entry.content && entry.content.length > 150 
								? entry.content.substring(0, 150) + '...' 
								: entry.content || ''}
						</p>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.line-clamp-3 {
		display: -webkit-box;
		-webkit-line-clamp: 3;
		line-clamp: 3;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>