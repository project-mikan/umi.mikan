<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import type { Entity } from "$lib/grpc/entity/entity_pb";

export let suggestions: Entity[] = [];
export let selectedIndex = -1;
export let onSelect: (entity: Entity) => void;
export let position: { top: number; left: number } = { top: 0, left: 0 };

// 候補選択
function handleSelect(entity: Entity) {
	onSelect(entity);
}

// キーボード操作のために外部から呼び出される
export function selectByIndex(index: number) {
	if (index >= 0 && index < suggestions.length) {
		handleSelect(suggestions[index]);
	}
}
</script>

{#if suggestions.length > 0}
	<div
		class="entity-suggestions"
		style="top: {position.top}px; left: {position.left}px;"
	>
		<ul class="suggestions-list">
			{#each suggestions as entity, i}
				<button
					type="button"
					class="suggestion-item {i === selectedIndex ? 'selected' : ''}"
					on:click={() => handleSelect(entity)}
					on:keydown={(e) => {
						if (e.key === 'Enter') {
							handleSelect(entity);
						}
					}}
				>
					<div class="entity-name">{entity.name}</div>
					{#if entity.aliases.length > 0}
						<div class="entity-aliases">
							{entity.aliases.map((a) => a.alias).join(', ')}
						</div>
					{/if}
				</button>
			{/each}
		</ul>
	</div>
{/if}

<style>
	.entity-suggestions {
		position: absolute;
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

	.suggestion-item {
		width: 100%;
		text-align: left;
		padding: 0.5rem;
		cursor: pointer;
		border: none;
		background: transparent;
		border-radius: 0.25rem;
		transition: background-color 0.15s;
	}

	.suggestion-item:hover,
	.suggestion-item.selected {
		background-color: #f3f4f6;
	}

	:global(.dark) .suggestion-item:hover,
	:global(.dark) .suggestion-item.selected {
		background-color: #4b5563;
	}

	.entity-name {
		font-weight: 500;
		color: #111827;
	}

	:global(.dark) .entity-name {
		color: #f9fafb;
	}

	.entity-aliases {
		font-size: 0.875rem;
		color: #6b7280;
		margin-top: 0.25rem;
	}

	:global(.dark) .entity-aliases {
		color: #9ca3af;
	}
</style>
