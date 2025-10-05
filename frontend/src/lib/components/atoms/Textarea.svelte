<script lang="ts">
import { createEventDispatcher, onMount, onDestroy } from "svelte";
import EntitySuggestions from "../molecules/EntitySuggestions.svelte";
import type { Entity } from "$lib/grpc/entity/entity_pb";
import type { DiaryEntityOutput } from "$lib/grpc/diary/diary_pb";
import { highlightEntities } from "$lib/utils/diary-entity-highlighter";

export let value = "";
export let placeholder = "";
export let required = false;
export let disabled = false;
export let id = "";
export let name = "";
export let rows = 4;
export let diaryEntities: DiaryEntityOutput[] = [];

// エンティティハイライトを適用すべき元のコンテンツ
// （保存されたコンテンツとdiaryEntitiesのpositionsが対応している）
let savedContent = "";
let previousDiaryEntities: DiaryEntityOutput[] = [];

const dispatch = createEventDispatcher();

// Entity候補関連
let suggestions: Entity[] = [];
// フラット化された候補リスト（エンティティ名とエイリアスを含む）
type FlatSuggestion = { entity: Entity; text: string; isAlias: boolean };
let flatSuggestions: FlatSuggestion[] = [];
let selectedSuggestionIndex = -1;
let showSuggestions = false;
let suggestionPosition = { top: 0, left: 0 };
let currentTriggerPos = -1;
let currentQuery = ""; // 現在の検索クエリ
let suggestionsComponent: EntitySuggestions;

let contentElement: HTMLDivElement;
let isUpdatingFromValue = false; // valueからの更新中かどうかのフラグ
let isComposing = false; // IME入力中かどうかのフラグ
let updateTimeout: ReturnType<typeof setTimeout> | null = null; // エンティティハイライト更新のタイムアウト
let isTyping = false; // ユーザーが入力中かどうかのフラグ

// diaryEntitiesが外部から変更されたら、保存されたコンテンツを更新
$: if (diaryEntities !== previousDiaryEntities) {
	previousDiaryEntities = diaryEntities;
	savedContent = value;
}

// エンティティハイライトを適用したHTMLを生成
// 入力中でない場合のみエンティティハイライトを表示
// また、現在のvalueが保存されたコンテンツと一致する場合のみハイライトを適用
$: highlightedHTML = (() => {
	// 入力中は常にプレーンテキスト
	if (isTyping) {
		return value.replace(/\n/g, "<br>");
	}

	// diaryEntitiesがない場合もプレーンテキスト
	if (!diaryEntities || diaryEntities.length === 0) {
		return value.replace(/\n/g, "<br>");
	}

	// 現在のvalueが保存されたコンテンツと異なる場合はプレーンテキスト
	// （編集中のテキストには古いpositionデータを適用しない）
	if (value !== savedContent) {
		return value.replace(/\n/g, "<br>");
	}

	// 保存されたコンテンツと一致する場合のみハイライトを適用
	return highlightEntities(value, diaryEntities);
})();

// captureフェーズでTabキーをキャプチャするためのリスナー
onMount(() => {
	if (contentElement) {
		// captureフェーズで追加してTabキーを早期にキャプチャ
		contentElement.addEventListener("keydown", _handleKeydown, true);
		// 初期値を設定
		updateContentElement();
	}
});

onDestroy(() => {
	if (contentElement) {
		contentElement.removeEventListener("keydown", _handleKeydown, true);
	}
});

const baseClasses =
	"block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md shadow-sm focus:outline-none resize-none min-h-24 whitespace-pre-wrap [&>br]:leading-none [&>br]:h-0";
$: classes = `${baseClasses} ${disabled ? "bg-gray-100 dark:bg-gray-800 cursor-not-allowed opacity-50" : ""}`;

// Calculate min height based on rows
$: minHeight = `${rows * 1.5}rem`;

// diaryEntitiesが外部から変更されたときのみコンテンツを更新
// 入力中や編集中のテキストには古いpositionデータを適用しない
$: if (
	contentElement &&
	!isUpdatingFromValue &&
	!isComposing &&
	!isTyping &&
	diaryEntities &&
	diaryEntities.length > 0 &&
	value === savedContent
) {
	updateContentElement();
}

function updateContentElement() {
	if (!contentElement) return;

	// 現在のカーソル位置を取得
	const selection = window.getSelection();
	let cursorPos = 0;

	if (selection && selection.rangeCount > 0) {
		const range = selection.getRangeAt(0);
		cursorPos = getTextOffset(
			contentElement,
			range.startContainer,
			range.startOffset,
		);
	}

	// HTMLを更新
	contentElement.innerHTML = highlightedHTML;

	// カーソル位置を復元
	if (cursorPos > 0) {
		restoreCursorPosition(cursorPos);
	}
}

function restoreCursorPosition(targetPos: number) {
	if (!contentElement) return;

	const selection = window.getSelection();
	if (!selection) return;

	let currentPos = 0;
	let targetNode: Node | null = null;
	let targetOffset = 0;

	function traverse(node: Node): boolean {
		if (node.nodeType === Node.TEXT_NODE) {
			const textLength = node.textContent?.length || 0;
			if (currentPos + textLength >= targetPos) {
				targetNode = node;
				targetOffset = targetPos - currentPos;
				return true;
			}
			currentPos += textLength;
		} else if (node.nodeType === Node.ELEMENT_NODE) {
			if (node.nodeName === "BR") {
				currentPos += 1;
				if (currentPos >= targetPos) {
					const parent = node.parentNode;
					if (parent) {
						targetNode = parent;
						targetOffset = Array.from(parent.childNodes).indexOf(
							node as ChildNode,
						);
						return true;
					}
				}
			}
			for (const child of Array.from(node.childNodes)) {
				if (traverse(child)) return true;
			}
		}
		return false;
	}

	traverse(contentElement);

	if (targetNode) {
		try {
			const range = document.createRange();
			// TypeScriptの型推論の問題を回避するため、Node型として明示的にキャスト
			const node = targetNode as Node;
			const textLength = node.textContent?.length || 0;
			range.setStart(node, Math.min(targetOffset, textLength));
			range.collapse(true);
			selection.removeAllRanges();
			selection.addRange(range);
		} catch (e) {
			// カーソル復元に失敗した場合は無視
		}
	}
}

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
	isUpdatingFromValue = true;
	isTyping = true; // 入力中フラグを立てる

	// 既存のタイムアウトをクリア
	if (updateTimeout !== null) {
		clearTimeout(updateTimeout);
	}

	value = htmlToPlainText(target.innerHTML);

	// contentElementが初期化されていない場合は何もしない
	if (!contentElement) {
		isUpdatingFromValue = false;
		return;
	}

	// Entity候補検索のロジック
	const selection = window.getSelection();
	let cursorPos = value.length; // デフォルトは末尾

	if (selection && selection.rangeCount > 0) {
		const range = selection.getRangeAt(0);
		cursorPos = getTextOffset(
			contentElement,
			range.startContainer,
			range.startOffset,
		);
	}

	const text = value;
	const beforeCursor = text.substring(0, cursorPos);

	// カーソル前の最後の単語を検索
	// 空白、改行、句読点で区切られた単語を抽出
	// エンティティ名には様々な文字が含まれる可能性があるため、
	// 区切り文字は最小限にする
	const entityMatch = beforeCursor.match(/([^\s\n。、！？,.!?]+)$/);
	if (entityMatch && entityMatch[1].length > 0) {
		const word = entityMatch[1];
		currentTriggerPos = cursorPos - word.length;
		currentQuery = word;
		await searchForSuggestions(word);
	} else {
		// エンティティ候補がない場合は候補を閉じる
		showSuggestions = false;
		currentQuery = "";
		currentTriggerPos = -1;
	}

	if (showSuggestions) {
		// カーソル位置に候補を表示
		updateSuggestionPosition(target);
	}

	isUpdatingFromValue = false;

	// 500ms後に入力が止まったら入力完了フラグを下ろす
	// （エンティティハイライトは保存後にサーバーから返されるpositionデータで適用される）
	updateTimeout = setTimeout(() => {
		isTyping = false; // 入力完了フラグ
	}, 500);
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

		// 候補をフラット化（エンティティ名とエイリアスを含む）
		flatSuggestions = [];
		for (const entity of suggestions) {
			// エンティティ名を追加
			flatSuggestions.push({ entity, text: entity.name, isAlias: false });
			// エイリアスを追加
			for (const alias of entity.aliases) {
				flatSuggestions.push({ entity, text: alias.alias, isAlias: true });
			}
		}

		showSuggestions = flatSuggestions.length > 0;
		// 候補が表示されたら最も先頭一致する候補を選択状態にする
		if (showSuggestions) {
			selectedSuggestionIndex = getBestFlatMatchIndex(query, flatSuggestions);
		}
	} catch (err) {
		console.error("Failed to search entities:", err);
		suggestions = [];
		flatSuggestions = [];
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

// 最も先頭一致する候補のインデックスを取得（フラット化されたリスト用）
function getBestFlatMatchIndex(
	query: string,
	flatSugs: FlatSuggestion[],
): number {
	if (!query || flatSugs.length === 0) return 0;

	const lowerQuery = query.toLowerCase();

	// 先頭一致するものを探す
	for (let i = 0; i < flatSugs.length; i++) {
		if (flatSugs[i].text.toLowerCase().startsWith(lowerQuery)) {
			return i;
		}
	}

	// 先頭一致がなければ最初の候補を返す
	return 0;
}

function _handleKeydown(event: KeyboardEvent) {
	// Entity候補のキーボード操作
	if (showSuggestions && flatSuggestions.length > 0) {
		// Tabキーは候補が表示されている時のみ処理
		if (event.key === "Tab") {
			event.preventDefault();
			event.stopPropagation();

			// Tabで次の候補へ、Shift+Tabで前の候補へ
			if (event.shiftKey) {
				selectedSuggestionIndex =
					selectedSuggestionIndex <= 0
						? flatSuggestions.length - 1
						: selectedSuggestionIndex - 1;
			} else {
				selectedSuggestionIndex =
					selectedSuggestionIndex >= flatSuggestions.length - 1
						? 0
						: selectedSuggestionIndex + 1;
			}

			return;
		}
		if (event.key === "ArrowDown") {
			event.preventDefault();
			selectedSuggestionIndex = Math.min(
				selectedSuggestionIndex + 1,
				flatSuggestions.length - 1,
			);
			return;
		}
		if (event.key === "ArrowUp") {
			event.preventDefault();
			selectedSuggestionIndex = Math.max(selectedSuggestionIndex - 1, -1);
			return;
		}
		if (event.key === "Enter") {
			event.preventDefault();
			event.stopPropagation();
			// 候補が選択されている場合はその候補を、
			// 選択されていない場合は最も先頭一致する候補を採用
			let indexToSelect: number;
			if (selectedSuggestionIndex >= 0) {
				indexToSelect = selectedSuggestionIndex;
			} else {
				indexToSelect = getBestFlatMatchIndex(currentQuery, flatSuggestions);
			}
			const selected = flatSuggestions[indexToSelect];
			selectSuggestion(selected.entity, selected.text);
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
function selectSuggestion(entity: Entity, selectedText?: string) {
	if (currentTriggerPos === -1) return;

	// valueのプレーンテキストから現在のカーソル位置を計算
	// （HTMLではなくvalueの文字数で計算）
	const selection = window.getSelection();
	if (!selection || selection.rangeCount === 0) return;

	const range = selection.getRangeAt(0);
	const cursorPos = getTextOffset(
		contentElement,
		range.startContainer,
		range.startOffset,
	);

	// 入力中の単語の開始位置
	const beforeTrigger = value.substring(0, currentTriggerPos);
	const afterCursor = value.substring(cursorPos);

	// 選択されたテキスト（エイリアスまたはエンティティ名）を使用
	const textToInsert = selectedText || entity.name;

	// 単語を選択されたテキストに置き換え（@とスペースなし）
	value = `${beforeTrigger}${textToInsert}${afterCursor}`;

	showSuggestions = false;
	selectedSuggestionIndex = -1;
	currentTriggerPos = -1;

	// 既存のタイムアウトをクリア
	if (updateTimeout !== null) {
		clearTimeout(updateTimeout);
	}

	// 入力中フラグを下ろす
	isTyping = false;

	// contentEditableの内容を更新してフォーカスを戻す
	setTimeout(() => {
		// プレーンテキスト（エンティティハイライトなし）で更新
		// （まだdiaryEntitiesに新しいエンティティ情報がないため）
		contentElement.innerHTML = value.replace(/\n/g, "<br>");
		contentElement.focus();

		// カーソル位置を設定（テキストの直後）
		const newCursorPos = beforeTrigger.length + textToInsert.length;
		const newRange = createRangeAtTextOffset(contentElement, newCursorPos);
		if (newRange) {
			const sel = window.getSelection();
			sel?.removeAllRanges();
			sel?.addRange(newRange);
		}
	}, 0);
}

// テキストオフセット位置にRangeを作成（contenteditable用）
function createRangeAtTextOffset(
	root: Node,
	targetOffset: number,
): Range | null {
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

function restoreCursorFromRange(range: Range) {
	const selection = window.getSelection();
	if (selection && range) {
		selection.removeAllRanges();
		selection.addRange(range);
	}
}

// Update content when value changes externally
// ただし、updateContentElement()が呼ばれる条件の場合はスキップ
// （エンティティハイライトが適用される場合は、updateContentElement()に任せる）
$: if (
	contentElement &&
	htmlToPlainText(contentElement.innerHTML) !== value &&
	!(
		!isUpdatingFromValue &&
		!isComposing &&
		!isTyping &&
		diaryEntities &&
		diaryEntities.length > 0 &&
		value === savedContent
	)
) {
	const savedRange = saveCursorPosition();
	contentElement.innerHTML = value.replace(/\n/g, "<br>");
	if (savedRange) {
		// Adjust range if it's out of bounds
		try {
			restoreCursorFromRange(savedRange);
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
		on:compositionstart={() => { isComposing = true; }}
		on:compositionend={() => { isComposing = false; }}
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
