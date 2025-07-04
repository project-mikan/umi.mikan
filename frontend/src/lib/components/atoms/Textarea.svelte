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
	"block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none resize-none min-h-24 whitespace-pre-wrap [&>br]:leading-none [&>br]:h-0";
$: classes = `${baseClasses} ${disabled ? "bg-gray-100 cursor-not-allowed" : ""}`;

// Calculate min height based on rows
$: minHeight = `${rows * 1.5}rem`;

function htmlToPlainText(html: string): string {
	// Create a temporary div to process HTML
	const tempDiv = document.createElement("div");
	tempDiv.innerHTML = html;

	// Convert common HTML elements to plain text
	// Replace <br> tags with newlines
	tempDiv.innerHTML = tempDiv.innerHTML.replace(/<br\s*\/?>/gi, "\n");

	// Replace <p> tags with newlines (Google Keep uses these)
	tempDiv.innerHTML = tempDiv.innerHTML.replace(/<\/p>/gi, "\n");
	tempDiv.innerHTML = tempDiv.innerHTML.replace(/<p[^>]*>/gi, "");

	// Replace <div> tags with newlines
	tempDiv.innerHTML = tempDiv.innerHTML.replace(/<\/div>/gi, "\n");
	tempDiv.innerHTML = tempDiv.innerHTML.replace(/<div[^>]*>/gi, "");

	// Replace list items with newlines and bullets
	tempDiv.innerHTML = tempDiv.innerHTML.replace(/<li[^>]*>/gi, "• ");
	tempDiv.innerHTML = tempDiv.innerHTML.replace(/<\/li>/gi, "\n");

	// Remove other common HTML tags while preserving content
	tempDiv.innerHTML = tempDiv.innerHTML.replace(
		/<\/?(?:ul|ol|strong|b|em|i|u|span|font)[^>]*>/gi,
		"",
	);

	// Get the plain text content
	let plainText = tempDiv.textContent || tempDiv.innerText || "";

	// Clean up extra whitespace and newlines only for pasted HTML content
	// Check if input contains complex HTML (not just br tags)
	const hasComplexHTML = /<(?!br\s*\/?>)[^>]+>/.test(html);

	if (hasComplexHTML) {
		// Only clean up pasted HTML content
		plainText = plainText
			.replace(/^\s+|\s+$/g, "") // Trim leading and trailing whitespace
			.replace(/[ \t]+/g, " "); // Replace multiple spaces/tabs with single space
	}

	return plainText;
}

function handleInput(event: Event) {
	const target = event.target as HTMLDivElement;
	value = htmlToPlainText(target.innerHTML);
}

function handleKeydown(event: KeyboardEvent) {
	if (event.ctrlKey && event.key === "Enter") {
		event.preventDefault();
		dispatch("save");
	} else if (event.key === "Enter") {
		// Prevent default behavior and manually insert <br>
		// This handles both Enter and Shift+Enter
		event.preventDefault();

		// Insert a <br> tag at the cursor position
		const selection = window.getSelection();
		if (selection && selection.rangeCount > 0) {
			const range = selection.getRangeAt(0);
			const br = document.createElement("br");

			// Delete any selected content first
			range.deleteContents();

			// Insert the br element
			range.insertNode(br);

			// Create a new range after the br element
			const newRange = document.createRange();
			newRange.setStartAfter(br);
			newRange.collapse(true);

			// Update the selection
			selection.removeAllRanges();
			selection.addRange(newRange);
		}

		// Trigger input event to update the value
		const inputEvent = new Event("input", { bubbles: true });
		contentElement.dispatchEvent(inputEvent);
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
$: if (contentElement && htmlToPlainText(contentElement.innerHTML) !== value) {
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
	style="min-height: {minHeight}; line-height: 18pt; font-size:11pt; font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre;white-space:pre-wrap;padding: 4px;"
	on:input={handleInput}
	on:keydown={handleKeydown}
	{...$$restProps}
></div>
