<script lang="ts">
import { createEventDispatcher } from "svelte";

export let value = "";
export let placeholder = "";
export let required = false;
export let disabled = false;
export let id = "";
export let name = "";
export let rows = 4;

const dispatch = createEventDispatcher();

let contentElement: HTMLDivElement;

const baseClasses =
	"block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 resize-none min-h-24 whitespace-pre-wrap [&>br]:leading-none [&>br]:h-0";
$: classes = `${baseClasses} ${disabled ? "bg-gray-100 cursor-not-allowed" : ""}`;

// Calculate min height based on rows
$: minHeight = `${rows * 1.5}rem`;

function handleInput(event: Event) {
	const target = event.target as HTMLDivElement;
	value =
		target.innerHTML
			.replace(/<br\s*\/?>/gi, "\n")
			.replace(/<div[^>]*>/gi, "")
			.replace(/<\/div>/gi, "\n")
			.replace(/^\n/, "")
			.replace(/\n$/g, "") || "";
}

function handleKeydown(event: KeyboardEvent) {
	if (event.ctrlKey && event.key === "Enter") {
		event.preventDefault();
		dispatch("save");
	}
}

function saveCursorPosition() {
	const selection = window.getSelection();
	if (selection && selection.rangeCount > 0) {
		return selection.getRangeAt(0);
	}
	return null;
}

function restoreCursorPosition(range: Range) {
	const selection = window.getSelection();
	if (selection && range) {
		selection.removeAllRanges();
		selection.addRange(range);
	}
}

// Update content when value changes externally
$: if (
	contentElement &&
	contentElement.innerHTML
		.replace(/<br\s*\/?>/gi, "\n")
		.replace(/<div[^>]*>/gi, "")
		.replace(/<\/div>/gi, "\n")
		.replace(/^\n/, "")
		.replace(/\n$/g, "") !== value
) {
	const savedRange = saveCursorPosition();
	contentElement.innerHTML = value.replace(/\n/g, "<br>");
	if (savedRange) {
		// Adjust range if it's out of bounds
		try {
			restoreCursorPosition(savedRange);
		} catch {
			// If range is invalid, place cursor at end
			const range = document.createRange();
			const selection = window.getSelection();
			range.selectNodeContents(contentElement);
			range.collapse(false);
			selection?.removeAllRanges();
			selection?.addRange(range);
		}
	}
}
</script>

<!-- Hidden input for form submission -->
<input type="hidden" {name} {value} {required} />

<div
	bind:this={contentElement}
	{id}
	data-placeholder={placeholder}
	contenteditable={!disabled}
	class={classes}
	style="min-height: {minHeight};"
	on:input={handleInput}
	on:keydown={handleKeydown}
	{...$$restProps}
></div>