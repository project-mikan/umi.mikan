<script lang="ts">
import { _ } from "svelte-i18n";
import { goto } from "$app/navigation";
import { enhance } from "$app/forms";
import "$lib/i18n";
import Modal from "$lib/components/molecules/Modal.svelte";
import { EntityCategory } from "$lib/grpc/entity/entity_pb";
import type { ActionData, PageData } from "./$types";

export let data: PageData;
export let form: ActionData;

// ローディング状態
let updateLoading = false;
let deleteLoading = false;
let createAliasLoading = false;
let deleteAliasLoading = false;

// モーダル状態
let showDeleteConfirm = false;
let showDeleteAliasConfirm = false;
let selectedAliasId = "";

// エイリアス追加用
let newAlias = "";

// アクションメッセージの判定
function isMessageForAction(actionName: string): boolean {
	return form?.action === actionName;
}

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
 * 日付フォーマット
 */
function formatDate(dateStr: string): string {
	const date = new Date(dateStr);
	return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`;
}

/**
 * 日記詳細ページへ遷移
 */
function viewDiary(dateStr: string): void {
	goto(`/${dateStr}`);
}

/**
 * エンティティ削除確認モーダルを表示
 */
function confirmDelete(): void {
	showDeleteConfirm = true;
}

/**
 * エンティティ削除をキャンセル
 */
function cancelDelete(): void {
	showDeleteConfirm = false;
}

/**
 * エンティティ削除を実行
 */
function handleDelete(): void {
	showDeleteConfirm = false;
	const form = document.createElement("form");
	form.method = "POST";
	form.action = "?/deleteEntity";
	document.body.appendChild(form);
	deleteLoading = true;
	form.submit();
}

/**
 * エイリアス削除確認モーダルを表示
 */
function confirmDeleteAlias(aliasId: string): void {
	selectedAliasId = aliasId;
	showDeleteAliasConfirm = true;
}

/**
 * エイリアス削除をキャンセル
 */
function cancelDeleteAlias(): void {
	showDeleteAliasConfirm = false;
	selectedAliasId = "";
}

/**
 * エイリアス削除を実行
 */
function handleDeleteAlias(): void {
	showDeleteAliasConfirm = false;
	const form = document.createElement("form");
	form.method = "POST";
	form.action = "?/deleteAlias";

	const input = document.createElement("input");
	input.type = "hidden";
	input.name = "aliasId";
	input.value = selectedAliasId;
	form.appendChild(input);

	document.body.appendChild(form);
	deleteAliasLoading = true;
	form.submit();
}
</script>

<svelte:head>
	<title>{data.entity?.name || $_("entity.detail.title")} - umi.mikan</title>
</svelte:head>

<div class="max-w-4xl mx-auto p-6">
	<!-- 戻るボタン -->
	<button
		on:click={() => goto("/entities")}
		class="mb-4 text-blue-600 dark:text-blue-400 hover:underline"
	>
		← {$_("entity.list.title")}
	</button>

	{#if data.entity}
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 mb-6">
			<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-4">
				{$_("entity.detail.title")}
			</h1>

			<!-- エンティティ情報編集フォーム -->
			<form
				method="POST"
				action="?/updateEntity"
				class="space-y-4"
				use:enhance={() => {
					updateLoading = true;
					return async ({ update }) => {
						updateLoading = false;
						await update();
					};
				}}
			>
				<div>
					<label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("entity.detail.name")}
					</label>
					<input
						type="text"
						id="name"
						name="name"
						required
						disabled={updateLoading}
						value={data.entity.name}
						class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 dark:disabled:bg-gray-800"
					/>
				</div>

				<div>
					<label for="category" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("entity.detail.category")}
					</label>
					<select
						id="category"
						name="category"
						disabled={updateLoading}
						value={data.entity.category}
						class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 dark:disabled:bg-gray-800"
					>
						<option value={EntityCategory.NO_CATEGORY}>{$_("entity.list.category.noCategory")}</option>
						<option value={EntityCategory.PEOPLE}>{$_("entity.list.category.people")}</option>
					</select>
				</div>

				<div>
					<label for="memo" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
						{$_("entity.detail.memo")}
					</label>
					<textarea
						id="memo"
						name="memo"
						rows="3"
						disabled={updateLoading}
						value={data.entity.memo}
						class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 dark:disabled:bg-gray-800"
					></textarea>
				</div>

				<div class="flex gap-2">
					<button
						type="submit"
						disabled={updateLoading}
						class="px-4 py-2 bg-blue-600 dark:bg-blue-500 text-white rounded-md hover:bg-blue-700 dark:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:bg-blue-300 dark:disabled:bg-blue-800"
					>
						{updateLoading ? $_("common.loading") : $_("entity.detail.update")}
					</button>
					<button
						type="button"
						on:click={confirmDelete}
						disabled={deleteLoading}
						class="px-4 py-2 bg-red-600 dark:bg-red-500 text-white rounded-md hover:bg-red-700 dark:hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 disabled:bg-red-300 dark:disabled:bg-red-800"
					>
						{deleteLoading ? $_("common.loading") : $_("entity.detail.delete")}
					</button>
				</div>

				<!-- 更新メッセージ -->
				{#if form?.error && isMessageForAction("updateEntity")}
					<div class="bg-red-100 dark:bg-red-900/20 border border-red-400 dark:border-red-600 text-red-700 dark:text-red-400 px-4 py-3 rounded auto-phrase-target">
						{$_(`entity.messages.${form.error}`) || form.error}
					</div>
				{/if}
				{#if form?.success && isMessageForAction("updateEntity")}
					<div class="bg-green-100 dark:bg-green-900/20 border border-green-400 dark:border-green-600 text-green-700 dark:text-green-400 px-4 py-3 rounded auto-phrase-target">
						{$_(`entity.messages.${form.message}`) || form.message}
					</div>
				{/if}
			</form>
		</div>

		<!-- エイリアス一覧 -->
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 mb-6">
			<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
				{$_("entity.detail.aliases")}
			</h2>

			<!-- エイリアス追加フォーム -->
			<form
				method="POST"
				action="?/createAlias"
				class="flex gap-2 mb-4"
				use:enhance={() => {
					createAliasLoading = true;
					return async ({ update }) => {
						createAliasLoading = false;
						newAlias = "";
						await update();
					};
				}}
			>
				<input
					type="text"
					name="alias"
					bind:value={newAlias}
					disabled={createAliasLoading}
					placeholder={$_("entity.detail.addAlias")}
					class="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 dark:disabled:bg-gray-800"
				/>
				<button
					type="submit"
					disabled={createAliasLoading || !newAlias}
					class="px-4 py-2 bg-blue-600 dark:bg-blue-500 text-white rounded-md hover:bg-blue-700 dark:hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:bg-blue-300 dark:disabled:bg-blue-800"
				>
					{createAliasLoading ? $_("common.loading") : $_("entity.detail.addAlias")}
				</button>
			</form>

			<!-- エイリアスメッセージ -->
			{#if form?.error && isMessageForAction("createAlias")}
				<div class="mb-4 bg-red-100 dark:bg-red-900/20 border border-red-400 dark:border-red-600 text-red-700 dark:text-red-400 px-4 py-3 rounded auto-phrase-target">
					{$_(`entity.messages.${form.error}`) || form.error}
				</div>
			{/if}
			{#if form?.success && isMessageForAction("createAlias")}
				<div class="mb-4 bg-green-100 dark:bg-green-900/20 border border-green-400 dark:border-green-600 text-green-700 dark:text-green-400 px-4 py-3 rounded auto-phrase-target">
					{$_(`entity.messages.${form.message}`) || form.message}
				</div>
			{/if}

			<!-- エイリアス一覧表示 -->
			{#if data.entity.aliases && data.entity.aliases.length > 0}
				<div class="flex flex-wrap gap-2">
					{#each data.entity.aliases as alias}
						<div class="flex items-center gap-2 px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 rounded-full">
							<span>{alias.alias}</span>
							<button
								type="button"
								on:click={() => confirmDeleteAlias(alias.id)}
								disabled={deleteAliasLoading}
								class="text-blue-700 dark:text-blue-400 hover:text-red-600 dark:hover:text-red-400 focus:outline-none"
								aria-label={$_("entity.detail.deleteAlias")}
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
								</svg>
							</button>
						</div>
					{/each}
				</div>
			{:else}
				<p class="text-gray-500 dark:text-gray-400 text-sm">
					{$_("entity.detail.addAlias")}
				</p>
			{/if}
		</div>

		<!-- 関連する日記 -->
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
			<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-4">
				{$_("entity.detail.relatedDiaries")}
			</h2>

			{#if data.diaries && data.diaries.length > 0}
				<div class="space-y-4">
					{#each data.diaries as diary}
						<div
							class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer transition-colors"
							on:click={() => viewDiary(diary.date)}
							on:keydown={(e) => e.key === "Enter" && viewDiary(diary.date)}
							role="button"
							tabindex="0"
						>
							<div class="flex justify-between items-start mb-2">
								<h3 class="font-semibold text-blue-600 dark:text-blue-400">
									{formatDate(diary.date)}
								</h3>
							</div>
							<p class="text-sm text-gray-600 dark:text-gray-400 line-clamp-2 auto-phrase-target">
								{diary.content.length > 150
									? diary.content.substring(0, 150) + "..."
									: diary.content}
							</p>
						</div>
					{/each}
				</div>
			{:else}
				<p class="text-gray-500 dark:text-gray-400 text-sm">
					{$_("entity.detail.noDiaries")}
				</p>
			{/if}
		</div>
	{/if}
</div>

<!-- エンティティ削除確認モーダル -->
<Modal
	isOpen={showDeleteConfirm}
	title={$_("entity.detail.deleteConfirm")}
	confirmText={$_("entity.detail.delete")}
	cancelText={$_("common.cancel")}
	variant="danger"
	onConfirm={handleDelete}
	onCancel={cancelDelete}
>
	<p class="text-sm text-gray-500 dark:text-gray-400 auto-phrase-target">
		{$_("entity.detail.deleteMessage")}
	</p>
</Modal>

<!-- エイリアス削除確認モーダル -->
<Modal
	isOpen={showDeleteAliasConfirm}
	title={$_("entity.detail.deleteAliasConfirm")}
	confirmText={$_("entity.detail.deleteAlias")}
	cancelText={$_("common.cancel")}
	variant="danger"
	onConfirm={handleDeleteAlias}
	onCancel={cancelDeleteAlias}
>
	<p class="text-sm text-gray-500 dark:text-gray-400 auto-phrase-target">
		{$_("entity.detail.deleteAliasMessage")}
	</p>
</Modal>

<style>
	.line-clamp-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
