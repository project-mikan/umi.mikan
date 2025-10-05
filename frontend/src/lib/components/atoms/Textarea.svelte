<script lang="ts">
import { createEventDispatcher } from "svelte";
import EntitySuggestions from "../molecules/EntitySuggestions.svelte";
import type { Entity } from "$lib/grpc/entity/entity_pb";

export let value = "";
export let placeholder = "";
export let required = false;
export let disabled = false;
export let id = "";
export let name = "";
export let rows = 4;

const dispatch = createEventDispatcher();

// Entity候補関連
let suggestions: Entity[] = [];
let selectedSuggestionIndex = -1;
let showSuggestions = false;
let suggestionPosition = { top: 0, left: 0 };
let currentTriggerPos = -1;
let suggestionsComponent: EntitySuggestions;

let contentElement: HTMLDivElement;

const baseClasses =
	"block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md shadow-sm focus:outline-none resize-none min-h-24 whitespace-pre-wrap [&>br]:leading-none [&>br]:h-0";
$: classes = `${baseClasses} ${disabled ? "bg-gray-100 dark:bg-gray-800 cursor-not-allowed opacity-50" : ""}`;

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

async function _handleInput(event: Event) {
	const target = event.target as HTMLDivElement;
	value = htmlToPlainText(target.innerHTML);

	// contentElementが初期化されていない場合は何もしない
	if (!contentElement) return;

	// Entity候補検索のロジック
	const selection = window.getSelection();
	let cursorPos = value.length; // デフォルトは末尾

	if (selection && selection.rangeCount > 0) {
		const range = selection.getRangeAt(0);
		cursorPos = getTextOffset(contentElement, range.startContainer, range.startOffset);
	}

	const text = value;

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
		updateSuggestionPosition(target);
	}
}

// テキスト位置を取得（contenteditable用）
function getTextOffset(root: Node, node: Node, offset: number): number {
	let textOffset = 0;
	const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);

	let currentNode = walker.nextNode();
	while (currentNode) {
		if (currentNode === node) {
			return textOffset + offset;
		}
		textOffset += currentNode.textContent?.length || 0;
		currentNode = walker.nextNode();
	}

	return textOffset;
}

// 候補検索
async function searchForSuggestions(query: string) {
	try {
		const response = await fetch(
			`/api/entities/search?q=${encodeURIComponent(query)}`,
		);
		const data = await response.json();
		suggestions = data.entities || [];
		showSuggestions = suggestions.length > 0;
	} catch (err) {
		console.error("Failed to search entities:", err);
		suggestions = [];
		showSuggestions = false;
	}
}

// 候補表示位置を更新
function updateSuggestionPosition(_target: HTMLDivElement) {
	// 親要素（relative div）からの相対位置で指定
	// contentElementの上端から少し下に表示
	suggestionPosition = {
		top: 25, // contentElementの上端から25px下
		left: 50, // 左端から50px右
	};
}

function _handleKeydown(event: KeyboardEvent) {
	// Entity候補のキーボード操作
	if (showSuggestions) {
		if (event.key === "ArrowDown") {
			event.preventDefault();
			selectedSuggestionIndex = Math.min(
				selectedSuggestionIndex + 1,
				suggestions.length - 1,
			);
			return;
		}
		if (event.key === "ArrowUp") {
			event.preventDefault();
			selectedSuggestionIndex = Math.max(selectedSuggestionIndex - 1, -1);
			return;
		}
		if (event.key === "Enter" && selectedSuggestionIndex >= 0) {
			event.preventDefault();
			selectSuggestion(suggestions[selectedSuggestionIndex]);
			return;
		}
		if (event.key === "Escape") {
			event.preventDefault();
			showSuggestions = false;
			selectedSuggestionIndex = -1;
			return;
		}
	}

	if (event.ctrlKey && event.key === "Enter") {
		event.preventDefault();
		dispatch("save");
	} else if (event.key === "Enter") {
		// Ignore Enter key during IME composition (Japanese input)
		if (event.isComposing) {
			return;
		}

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

			// Check if we're at the end of the content
			const isAtEnd =
				range.endContainer === contentElement &&
				range.endOffset === contentElement.childNodes.length;

			// Check if we're at the end of content or at the end of a text node
			const isAtEndOfContent =
				isAtEnd ||
				(range.endContainer.nodeType === Node.TEXT_NODE &&
					range.endOffset === range.endContainer.textContent?.length);

			if (isAtEndOfContent) {
				// For the last line, we need to insert a text node to position the cursor properly
				const textNode = document.createTextNode("");
				range.insertNode(textNode);

				// Position cursor after the br and before the text node
				const newRange = document.createRange();
				newRange.setStartAfter(br);
				newRange.setEndBefore(textNode);
				newRange.collapse(false);

				selection.removeAllRanges();
				selection.addRange(newRange);
			} else {
				// Create a new range after the br element
				const newRange = document.createRange();
				newRange.setStartAfter(br);
				newRange.collapse(true);

				// Update the selection
				selection.removeAllRanges();
				selection.addRange(newRange);
			}
		}

		// Trigger input event to update the value
		const inputEvent = new Event("input", { bubbles: true });
		contentElement.dispatchEvent(inputEvent);
	}
}

// 候補選択
function selectSuggestion(entity: Entity) {
	if (currentTriggerPos === -1) return;

	// @記号から現在のカーソル位置までを entity の名前に置き換え
	const selection = window.getSelection();
	if (!selection || selection.rangeCount === 0) return;

	const range = selection.getRangeAt(0);
	const cursorPos = getTextOffset(contentElement, range.startContainer, range.startOffset);

	const beforeTrigger = value.substring(0, currentTriggerPos);
	const afterCursor = value.substring(cursorPos);
	value = `${beforeTrigger}@${entity.name} ${afterCursor}`;

	showSuggestions = false;
	selectedSuggestionIndex = -1;
	currentTriggerPos = -1;

	// contentEditableの内容を更新してフォーカスを戻す
	setTimeout(() => {
		contentElement.innerHTML = value.replace(/\n/g, "<br>");
		contentElement.focus();

		// カーソル位置を設定
		const newCursorPos = beforeTrigger.length + entity.name.length + 2; // @ + name + space
		const newRange = createRangeAtTextOffset(contentElement, newCursorPos);
		if (newRange) {
			const sel = window.getSelection();
			sel?.removeAllRanges();
			sel?.addRange(newRange);
		}
	}, 0);
}

// テキストオフセット位置にRangeを作成（contenteditable用）
function createRangeAtTextOffset(root: Node, targetOffset: number): Range | null {
	const range = document.createRange();
	const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);

	let currentOffset = 0;
	let currentNode = walker.nextNode();

	while (currentNode) {
		const nodeLength = currentNode.textContent?.length || 0;
		if (currentOffset + nodeLength >= targetOffset) {
			const offset = targetOffset - currentOffset;
			range.setStart(currentNode, offset);
			range.collapse(true);
			return range;
		}
		currentOffset += nodeLength;
		currentNode = walker.nextNode();
	}

	// オフセットが範囲外の場合は最後に設定
	if (root.lastChild) {
		range.setStartAfter(root.lastChild);
		range.collapse(true);
		return range;
	}

	return null;
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

<div class="relative">
	<div
		bind:this={contentElement}
		{id}
		data-placeholder={placeholder}
		contenteditable={!disabled}
		class="{classes} auto-phrase-target"
		style="min-height: {minHeight}; line-height: 18pt; font-size:11pt; font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre;white-space:pre-wrap;padding: 4px;"
		on:input={_handleInput}
		on:keydown={_handleKeydown}
		{...$$restProps}
	></div>
	{#if showSuggestions}
		<EntitySuggestions
			bind:this={suggestionsComponent}
			{suggestions}
			selectedIndex={selectedSuggestionIndex}
			position={suggestionPosition}
			onSelect={selectSuggestion}
		/>
	{/if}
</div>
