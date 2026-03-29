<script lang="ts">
	import { createEventDispatcher, onMount, onDestroy } from "svelte";
	import { highlightEntitiesAndHighlights } from "$lib/utils/diary-entity-highlighter";
	import type { DiaryHighlight } from "$lib/types/highlight";
	import {
		getTextOffset,
		restoreCursorPosition,
		saveCursorPosition,
		restoreCursorFromRange,
	} from "$lib/utils/cursor-utils";
	import { htmlToPlainText } from "$lib/utils/html-text-converter";

	export let value = "";
	export let placeholder = "";
	export let required = false;
	export let disabled = false;
	export let id = "";
	export let name = "";
	export let rows = 4;
	export let diaryHighlights: DiaryHighlight[] = [];

	// ハイライトを適用すべき元のコンテンツ
	let savedContent = "";

	const dispatch = createEventDispatcher();

	let contentElement: HTMLDivElement;
	let isUpdatingFromValue = false; // valueからの更新中かどうかのフラグ
	let isComposing = false; // IME入力中かどうかのフラグ
	let updateTimeout: ReturnType<typeof setTimeout> | null = null; // ハイライト更新のタイムアウト
	let isTyping = false; // ユーザーが入力中かどうかのフラグ

	// diaryHighlightsが変更されたときもsavedContentを更新
	// 配列の内容をシリアライズして比較（参照比較では毎回新しい配列なので不十分）
	let previousDiaryHighlightsKey = "";
	$: {
		// valueとdiaryHighlightsへの依存を明示（初期化タイミングを確実にする）
		void value;

		// diaryHighlightsの内容をキーに変換（長さ + 最初の要素の情報）
		const currentKey = diaryHighlights
			? `${diaryHighlights.length}-${diaryHighlights[0]?.start ?? ""}-${diaryHighlights[0]?.end ?? ""}`
			: "empty";

		if (currentKey !== previousDiaryHighlightsKey) {
			previousDiaryHighlightsKey = currentKey;
			savedContent = value;
		}
	}

	// ハイライトを適用したHTMLを生成
	// 入力中でない場合のみハイライトを表示
	// また、現在のvalueが保存されたコンテンツと一致する場合のみハイライトを適用
	$: highlightedHTML = (() => {
		// 入力中は常にプレーンテキスト
		if (isTyping) {
			return value.replace(/\n/g, "<br>");
		}

		// diaryHighlightsがない場合はプレーンテキスト
		if (!diaryHighlights || diaryHighlights.length === 0) {
			return value.replace(/\n/g, "<br>");
		}

		// 現在のvalueが保存されたコンテンツと異なる場合はプレーンテキスト
		// （編集中のテキストには古いpositionデータを適用しない）
		if (value !== savedContent) {
			return value.replace(/\n/g, "<br>");
		}

		// diaryHighlightsがある場合はハイライトを適用
		return highlightEntitiesAndHighlights(value, [], diaryHighlights);
	})();

	// captureフェーズでTabキーをキャプチャするためのリスナー
	onMount(async () => {
		// savedContentを確実に初期化（SSR時のリアクティビティの問題を回避）
		if (!savedContent) {
			savedContent = value;
		}

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

	// diaryHighlightsが外部から変更されたときもコンテンツを更新
	// 空配列の場合もハイライトを消すために更新が必要
	// 配列の長さを監視して変更を確実に検知
	let previousHighlightsLength = -1;
	$: {
		// contentElementが存在する場合のみ処理
		if (contentElement && !isUpdatingFromValue && !isComposing && !isTyping) {
			// diaryHighlightsの長さを取得（nullまたはundefinedの場合は0）
			const currentLength = diaryHighlights ? diaryHighlights.length : 0;
			// 長さが変わった場合のみ更新（初回は-1から0以上に変わるので必ず更新される）
			if (currentLength !== previousHighlightsLength) {
				previousHighlightsLength = currentLength;
				updateContentElement();
			}
		}
	}

	function updateContentElement() {
		if (!contentElement) return;
		// SSR時は何もしない
		if (typeof window === "undefined") return;

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
			restoreCursorPosition(contentElement, cursorPos);
		}
	}

	function _handleInput(event: Event) {
		const target = event.target as HTMLDivElement;
		isUpdatingFromValue = true;
		isTyping = true; // 入力中フラグを立てる

		// 既存のタイムアウトをクリア
		if (updateTimeout !== null) {
			clearTimeout(updateTimeout);
		}

		// IME入力中（compositionupdate）の場合は、valueの更新のみ行う
		const isCompositionUpdate = event instanceof CompositionEvent;

		value = htmlToPlainText(target.innerHTML);

		// contentElementが初期化されていない場合は何もしない
		if (!contentElement) {
			isUpdatingFromValue = false;
			return;
		}

		if (isCompositionUpdate) {
			isUpdatingFromValue = false;

			// 500ms後に入力が止まったら入力完了フラグを下ろす
			updateTimeout = setTimeout(() => {
				isTyping = false;
			}, 500);

			return;
		}

		isUpdatingFromValue = false;

		// 500ms後に入力が止まったら入力完了フラグを下ろす
		// （ハイライトは保存後にサーバーから返されるpositionデータで適用される）
		updateTimeout = setTimeout(() => {
			isTyping = false; // 入力完了フラグ
		}, 500);
	}

	function _handleKeydown(event: KeyboardEvent) {
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

				// カーソルの直前のノードをチェック
				// 直前のノードがBRタグの場合、連続改行とみなす
				let previousNode = range.startContainer;
				if (
					previousNode.nodeType === Node.TEXT_NODE &&
					range.startOffset === 0
				) {
					// テキストノードの先頭にいる場合、前の兄弟ノードをチェック
					previousNode = previousNode.previousSibling as Node;
				} else if (previousNode.nodeType === Node.ELEMENT_NODE) {
					// 要素ノード内にいる場合、startOffsetで指定された位置の直前の子ノードをチェック
					const children = previousNode.childNodes;
					if (range.startOffset > 0 && range.startOffset <= children.length) {
						previousNode = children[range.startOffset - 1];
					}
				}

				const isPreviousNodeBR = previousNode && previousNode.nodeName === "BR";

				// Insert the br element
				range.insertNode(br);

				// contentEditableの末尾での改行を処理するために、
				// <br>の後ろに意味のあるコンテンツが存在するかチェック
				// 末尾の場合のみ、もう一つ<br>を挿入してカーソルが次の行に表示されるようにする
				// ただし、直前のノードがBRタグの場合（連続改行）は、1つのbrのみを挿入
				function hasContentAfterNode(node: Node): boolean {
					let current: Node | null = node.nextSibling;
					while (current) {
						// テキストノードの場合
						if (current.nodeType === Node.TEXT_NODE) {
							const text = current.textContent || "";
							// 空白以外のテキストがある場合はコンテンツありとみなす
							if (text.trim().length > 0) {
								return true;
							}
							// 空白のみの場合は次のノードをチェック
						}
						// 要素ノードの場合
						else if (current.nodeType === Node.ELEMENT_NODE) {
							const element = current as Element;
							// BRタグ以外の要素があればコンテンツありとみなす
							if (element.nodeName !== "BR") {
								return true;
							}
							// BRタグの場合は次のノードをチェック（連続するBRタグを全て無視）
						}
						current = current.nextSibling;
					}
					// 親ノードがcontentElementでない場合、親の兄弟をチェック
					if (node.parentNode && node.parentNode !== contentElement) {
						return hasContentAfterNode(node.parentNode);
					}
					return false;
				}

				const isAtEnd = !hasContentAfterNode(br);

				// 直前のノードがBRタグでない かつ 末尾の場合のみ、2つ目の<br>を挿入
				if (!isPreviousNodeBR && isAtEnd) {
					// 末尾の場合のみ2つ目の<br>を挿入
					const afterBr = document.createElement("br");

					// カーソルを最初の<br>の直後に配置してから2つ目の<br>を挿入
					const newRange = document.createRange();
					newRange.setStartAfter(br);
					newRange.collapse(true);

					// 2つ目の<br>を挿入
					newRange.insertNode(afterBr);

					// カーソルを2つの<br>の間に配置
					newRange.setStartAfter(br);
					newRange.setEndBefore(afterBr);
					newRange.collapse(true);

					// Update the selection
					selection.removeAllRanges();
					selection.addRange(newRange);
				} else {
					// 連続改行の場合、または末尾でない場合は1つの<br>のみで、カーソルをその直後に配置
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

			// 改行後はすぐに入力完了フラグを下ろして、カーソル位置が戻されないようにする
			// 既存のタイムアウトをクリア
			if (updateTimeout !== null) {
				clearTimeout(updateTimeout);
			}
			isTyping = false;

			// savedContentを更新して、ハイライト復元を防ぐ
			// （改行直後は新しいコンテンツとして扱う）
			savedContent = "";
		}
	}

	// フォーカスが外れた時の自動保存
	function _handleBlur() {
		// IME入力中は自動保存しない
		if (isComposing) return;
		dispatch("autosave");
	}

	// Update content when value changes externally
	// ただし、updateContentElement()が呼ばれる条件の場合はスキップ
	$: if (
		contentElement &&
		htmlToPlainText(contentElement.innerHTML) !== value
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
		on:blur={_handleBlur}
		on:compositionstart={() => { isComposing = true; }}
		on:compositionupdate={_handleInput}
		on:compositionend={(event) => {
			isComposing = false;
			// IME確定後に処理を実行
			_handleInput(event);
		}}
		{...$$restProps}
	></div>
</div>
