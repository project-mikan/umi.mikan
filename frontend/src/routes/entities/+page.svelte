<script lang="ts">
import { _ } from "svelte-i18n";
import { goto } from "$app/navigation";
import "$lib/i18n";
import { EntityCategory } from "$lib/grpc/entity/entity_pb";
import type { PageData } from "./$types";

export let data: PageData;

/**
 * カテゴリ名を表示用に変換
 */
function getCategoryLabel(category: EntityCategory): string {
	switch (category) {
		case EntityCategory.PEOPLE:
			return $_("entity.list.category.people");
		case EntityCategory.NO_CATEGORY:
			return $_("entity.list.category.noCategory");
		default:
			return $_("entity.list.category.noCategory");
	}
}

/**
 * エンティティ詳細ページへ遷移
 */
function viewEntity(id: string): void {
	goto(`/entity/${id}`);
}

/**
 * カテゴリフィルタを変更
 */
function changeCategory(category: string): void {
	goto(`/entities?category=${category}`);
}
</script>

<svelte:head>
	<title>{$_("entity.list.title")} - umi.mikan</title>
</svelte:head>

<div class="max-w-6xl mx-auto p-6">
	<div class="flex justify-between items-center mb-6">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">
			{$_("entity.list.title")}
		</h1>
		<button
			on:click={() => goto("/entity/new")}
			class="px-4 py-2 bg-blue-600 dark:bg-blue-500 text-white rounded-lg hover:bg-blue-700 dark:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
		>
			{$_("entity.list.create")}
		</button>
	</div>

	<!-- カテゴリフィルタ -->
	<div class="mb-6 flex gap-2">
		<button
			on:click={() => changeCategory("all")}
			class="px-4 py-2 rounded-lg {data.selectedCategory === 'all'
				? 'bg-blue-600 dark:bg-blue-500 text-white'
				: 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600'} hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-blue-500"
		>
			{$_("entity.list.category.all")}
		</button>
		<button
			on:click={() => changeCategory("people")}
			class="px-4 py-2 rounded-lg {data.selectedCategory === 'people'
				? 'bg-blue-600 dark:bg-blue-500 text-white'
				: 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600'} hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-blue-500"
		>
			{$_("entity.list.category.people")}
		</button>
		<button
			on:click={() => changeCategory("noCategory")}
			class="px-4 py-2 rounded-lg {data.selectedCategory === 'noCategory'
				? 'bg-blue-600 dark:bg-blue-500 text-white'
				: 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600'} hover:opacity-80 focus:outline-none focus:ring-2 focus:ring-blue-500"
		>
			{$_("entity.list.category.noCategory")}
		</button>
	</div>

	<!-- エラーメッセージ -->
	{#if data.error}
		<div class="bg-red-100 dark:bg-red-900/20 border border-red-400 dark:border-red-600 text-red-700 dark:text-red-400 px-4 py-3 rounded mb-4">
			{data.error}
		</div>
	{/if}

	<!-- エンティティ一覧 -->
	{#if data.entities && data.entities.length > 0}
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each data.entities as entity}
				<div
					class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 shadow-sm dark:shadow-gray-900/20 hover:shadow-md dark:hover:shadow-gray-900/30 transition-shadow cursor-pointer"
					on:click={() => viewEntity(entity.id)}
					on:keydown={(e) => e.key === "Enter" && viewEntity(entity.id)}
					role="button"
					tabindex="0"
				>
					<div class="flex justify-between items-start mb-2">
						<h3 class="text-lg font-semibold text-blue-600 dark:text-blue-400">
							{entity.name}
						</h3>
						<span class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400">
							{getCategoryLabel(entity.category)}
						</span>
					</div>
					{#if entity.memo}
						<p class="text-sm text-gray-600 dark:text-gray-400 line-clamp-2 mb-2 auto-phrase-target">
							{entity.memo}
						</p>
					{/if}
					{#if entity.aliases && entity.aliases.length > 0}
						<div class="flex flex-wrap gap-1 mt-2">
							{#each entity.aliases as alias}
								<span class="text-xs px-2 py-0.5 rounded bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400">
									{alias.alias}
								</span>
							{/each}
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{:else}
		<div class="text-center py-8 text-gray-500 dark:text-gray-400">
			<p>{$_("entity.list.empty")}</p>
		</div>
	{/if}
</div>

<style>
	.line-clamp-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
