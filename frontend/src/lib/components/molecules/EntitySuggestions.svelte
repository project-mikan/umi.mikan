<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import type { Entity, EntityAlias } from "$lib/grpc/entity/entity_pb";

// フラット化された候補リスト
type FlatSuggestion = { entity: Entity; text: string; isAlias: boolean };

export let flatSuggestions: FlatSuggestion[] = [];
export let selectedIndex = -1;
export let onSelect: (entity: Entity, selectedText?: string) => void;
export let position: { top: number; left: number } = { top: 0, left: 0 };

// 候補選択
function handleSelect(entity: Entity, selectedText?: string) {
	onSelect(entity, selectedText);
}

// キーボード操作のために外部から呼び出される
export function selectByIndex(index: number) {
	if (index >= 0 && index < flatSuggestions.length) {
		const selected = flatSuggestions[index];
		handleSelect(selected.entity, selected.text);
	}
}
</script>

{#if flatSuggestions.length > 0}
	<div
		class="entity-suggestions"
		style="top: {position.top}px; left: {position.left}px;"
	>
		<ul class="suggestions-list">
			{#each flatSuggestions as suggestion, i}
				<li>
					<button
						type="button"
						class="suggestion-item {i === selectedIndex ? 'selected' : ''} {suggestion.isAlias ? 'alias' : ''}"
						on:click={() => handleSelect(suggestion.entity, suggestion.text)}
						on:keydown={(e) => {
							if (e.key === 'Enter') {
								e.preventDefault();
								e.stopPropagation();
								handleSelect(suggestion.entity, suggestion.text);
							}
						}}
					>
						{#if suggestion.isAlias}
							<span class="alias-prefix">→</span>
						{/if}
						{suggestion.text}
					</button>
				</li>
			{/each}
		</ul>
	</div>
{/if}

<style>
	.entity-suggestions {
		position: fixed;
		z-index: 1000;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.375rem;
		box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
		max-height: 200px;
		overflow-y: auto;
		min-width: 200px;
	}

	:global(.dark) .entity-suggestions {
		background: #374151;
		border-color: #4b5563;
	}

	.suggestions-list {
		list-style: none;
		margin: 0;
		padding: 0.25rem;
	}

	.suggestions-list li {
		list-style: none;
	}

	.suggestion-item {
		width: 100%;
		text-align: left;
		padding: 0.5rem;
		cursor: pointer;
		border: none;
		background: transparent;
		border-radius: 0.25rem;
		transition: background-color 0.15s;
		font-weight: 500;
		color: #111827;
	}

	.suggestion-item.alias {
		font-weight: 400;
		color: #6b7280;
		padding-left: 1.5rem;
	}

	.suggestion-item:hover,
	.suggestion-item.selected {
		background-color: #f3f4f6;
	}

	:global(.dark) .suggestion-item {
		color: #f9fafb;
	}

	:global(.dark) .suggestion-item.alias {
		color: #9ca3af;
	}

	:global(.dark) .suggestion-item:hover,
	:global(.dark) .suggestion-item.selected {
		background-color: #4b5563;
	}

	.alias-prefix {
		margin-right: 0.5rem;
		color: #9ca3af;
	}
</style>
