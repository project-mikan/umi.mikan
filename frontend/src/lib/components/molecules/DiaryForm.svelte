<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import { enhance } from "$app/forms";
import { goto } from "$app/navigation";
import Alert from "../atoms/Alert.svelte";
import Button from "../atoms/Button.svelte";
import FormField from "./FormField.svelte";
import EntitySuggestions from "./EntitySuggestions.svelte";
import type { Entity } from "$lib/grpc/entity/entity_pb";

export let title: string;
export let content = "";
export let date = "";
export let error: string | undefined = undefined;
export let showDeleteButton = false;
export let onCancel: (() => void) | null = null;
export let onDelete: (() => void) | null = null;

let isSubmitting = false;
let textarea: HTMLTextAreaElement;
type FlatSuggestion = { entity: Entity; text: string; isAlias: boolean };
let flatSuggestions: FlatSuggestion[] = [];
let selectedSuggestionIndex = -1;
let showSuggestions = false;
let suggestionPosition = { top: 0, left: 0 };
let currentTriggerPos = -1; // @記号の位置
let suggestionsComponent: EntitySuggestions;

function _handleCancel() {
	if (onCancel) {
		onCancel();
	} else {
		goto("/");
	}
}

// 候補検索
async function searchForSuggestions(query: string) {
	try {
		// 空文字列の場合は全候補を取得するため、空文字列でも検索
		const response = await fetch(
			`/api/entities/search?q=${encodeURIComponent(query)}`,
		);
		const data = await response.json();
		const entities: Entity[] = data.entities || [];

		// フラット化して保存
		flatSuggestions = [];
		for (const entity of entities) {
			// エンティティ名を追加
			flatSuggestions.push({ entity, text: entity.name, isAlias: false });
			// エイリアスを追加
			for (const alias of entity.aliases) {
				flatSuggestions.push({ entity, text: alias.alias, isAlias: true });
			}
		}

		showSuggestions = flatSuggestions.length > 0;
	} catch (err) {
		console.error("Failed to search entities:", err);
		flatSuggestions = [];
		showSuggestions = false;
	}
}

// テキストエリアでの入力処理
async function handleInput(event: Event) {
	const target = event.target as HTMLTextAreaElement;
	const cursorPos = target.selectionStart;
	const text = target.value;

	// @記号を探す
	const beforeCursor = text.substring(0, cursorPos);
	const lastAtPos = beforeCursor.lastIndexOf("@");

	if (lastAtPos === -1) {
		showSuggestions = false;
		return;
	}

	// @記号の後の文字列を取得（スペースまたは改行があれば終了）
	const afterAt = beforeCursor.substring(lastAtPos + 1);
	if (/[\s\n]/.test(afterAt)) {
		showSuggestions = false;
		return;
	}

	// 候補を検索
	currentTriggerPos = lastAtPos;
	await searchForSuggestions(afterAt);

	if (showSuggestions) {
		// カーソル位置に候補を表示
		updateSuggestionPosition(target, cursorPos);
	}
}

// 候補表示位置を更新
function updateSuggestionPosition(
	target: HTMLTextAreaElement,
	_cursorPos: number,
) {
	// 簡易的な位置計算（実際のカーソル位置を取得するのは複雑なので、textareaの下に表示）
	const rect = target.getBoundingClientRect();
	suggestionPosition = {
		top: rect.bottom + window.scrollY,
		left: rect.left + window.scrollX,
	};
}

// キーボード操作
function handleKeyDown(event: KeyboardEvent) {
	if (!showSuggestions) return;

	if (event.key === "ArrowDown") {
		event.preventDefault();
		selectedSuggestionIndex = Math.min(
			selectedSuggestionIndex + 1,
			flatSuggestions.length - 1,
		);
	} else if (event.key === "ArrowUp") {
		event.preventDefault();
		selectedSuggestionIndex = Math.max(selectedSuggestionIndex - 1, -1);
	} else if (event.key === "Enter" && selectedSuggestionIndex >= 0) {
		event.preventDefault();
		event.stopPropagation();
		const selected = flatSuggestions[selectedSuggestionIndex];
		selectSuggestion(selected.entity, selected.text);
	} else if (event.key === "Escape") {
		event.preventDefault();
		showSuggestions = false;
		selectedSuggestionIndex = -1;
	}
}

// 候補選択
function selectSuggestion(entity: Entity, selectedText?: string) {
	if (currentTriggerPos === -1) return;

	const textToInsert = selectedText || entity.name;

	// @記号から現在のカーソル位置までを選択されたテキストに置き換え
	const beforeTrigger = content.substring(0, currentTriggerPos);
	const afterCursor = content.substring(textarea.selectionStart);
	content = `${beforeTrigger}@${textToInsert} ${afterCursor}`;

	showSuggestions = false;
	selectedSuggestionIndex = -1;
	currentTriggerPos = -1;

	// フォーカスを戻す
	setTimeout(() => {
		textarea.focus();
		const newCursorPos = beforeTrigger.length + textToInsert.length + 2; // @ + text + space
		textarea.setSelectionRange(newCursorPos, newCursorPos);
	}, 0);
}
</script>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">{title}</h1>
		<Button
			variant="secondary"
			size="md"
			on:click={_handleCancel}
		>
			{$_('diary.back')}
		</Button>
	</div>

	<div class="bg-white dark:bg-gray-800 shadow dark:shadow-gray-900/20 rounded-lg p-6">
		<form method="POST" use:enhance={() => {
			isSubmitting = true;
			return async ({ result }) => {
				isSubmitting = false;
			};
		}}>
			{#if error}
				<Alert type="error">
					{error}
				</Alert>
			{/if}

			<FormField
				type="input"
				inputType="date"
				label={$_('create.date')}
				id="date"
				name="date"
				required
				bind:value={date}
			/>

			<div class="mb-6 relative">
				<label
					for="content"
					class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
				>
					{$_('create.content')}
				</label>
				<textarea
					id="content"
					name="content"
					bind:this={textarea}
					bind:value={content}
					on:input={handleInput}
					on:keydown={handleKeyDown}
					placeholder={$_('diary.placeholder')}
					required
					rows={12}
					class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:text-gray-100"
				></textarea>
				{#if showSuggestions}
					<EntitySuggestions
						bind:this={suggestionsComponent}
						{flatSuggestions}
						selectedIndex={selectedSuggestionIndex}
						position={suggestionPosition}
						onSelect={selectSuggestion}
					/>
				{/if}
			</div>

			<div class="flex {showDeleteButton ? 'justify-between' : 'justify-end'}">
				{#if showDeleteButton && onDelete}
					<Button
						type="button"
						variant="danger"
						size="md"
						on:click={onDelete}
					>
						{$_('diary.delete')}
					</Button>
				{/if}

				<div class="flex space-x-4">
					<Button
						type="button"
						variant="secondary"
						size="md"
						on:click={_handleCancel}
					>
						{$_('diary.cancel')}
					</Button>
					<Button
						type="submit"
						variant="primary"
						size="md"
						disabled={isSubmitting}
					>
						{#if isSubmitting}
							<div class="flex items-center">
								<svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								{$_('diary.saving')}
							</div>
						{:else}
							{$_('diary.save')}
						{/if}
					</Button>
				</div>
			</div>
		</form>
	</div>
</div>