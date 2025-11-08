<script lang="ts">
	import { createEventDispatcher, onMount, onDestroy, tick } from "svelte";
	import EntitySuggestions from "../molecules/EntitySuggestions.svelte";
	import type { Entity } from "$lib/grpc/entity/entity_pb";
	import type { DiaryEntityOutput } from "$lib/grpc/diary/diary_pb";
	import {
		highlightEntities,
		highlightEntitiesAndHighlights,
		validateDiaryEntities,
		type DiaryHighlight,
	} from "$lib/utils/diary-entity-highlighter";
	import {
		getTextOffset,
		restoreCursorPosition,
		createRangeAtTextOffset,
		saveCursorPosition,
		restoreCursorFromRange,
	} from "$lib/utils/cursor-utils";
	import { htmlToPlainText } from "$lib/utils/html-text-converter";
	import {
		type FlatSuggestion,
		type SelectedEntity,
		loadAllEntities,
		filterEntitiesByPrefix,
		getBestMatchIndex,
		findLongestMatch,
		adjustPositions,
		generateEntityHighlightHTML,
		syncSelectedEntitiesFromDOM,
	} from "$lib/utils/entity-completion";

	export let value = "";
	export let placeholder = "";
	export let required = false;
	export let disabled = false;
	export let id = "";
	export let name = "";
	export let rows = 4;
	export let diaryEntities: DiaryEntityOutput[] = [];
	export let diaryHighlights: DiaryHighlight[] = [];

	// 明示的に選択されたエンティティの情報を格納
	export let selectedEntities: SelectedEntity[] = [];

	// エンティティハイライトを適用すべき元のコンテンツ
	// （保存されたコンテンツとdiaryEntitiesのpositionsが対応している）
	let savedContent = "";
	let previousDiaryEntities: DiaryEntityOutput[] = [];

	const dispatch = createEventDispatcher();

	// Entity候補関連
	let flatSuggestions: FlatSuggestion[] = [];
	let selectedSuggestionIndex = -1;
	let showSuggestions = false;
	let suggestionPosition = { top: 0, left: 0 };
	let currentTriggerPos = -1;
	let currentQuery = ""; // 現在の検索クエリ
	let suggestionsComponent: EntitySuggestions;
	let justSelectedEntity = false; // エンティティを確定した直後かどうか
	let lastSelectedEntityText = ""; // 最後に確定したエンティティ名
	let isSelectingEntity = false; // エンティティ選択処理中かどうか（reactive statementからの上書きを防ぐ）

	// 全エンティティデータをキャッシュ（ページロード時に取得）
	let allEntities: Entity[] = [];
	let allFlatEntities: FlatSuggestion[] = [];

	let contentElement: HTMLDivElement;
	let isUpdatingFromValue = false; // valueからの更新中かどうかのフラグ
	let isComposing = false; // IME入力中かどうかのフラグ
	let updateTimeout: ReturnType<typeof setTimeout> | null = null; // エンティティハイライト更新のタイムアウト
	let isTyping = false; // ユーザーが入力中かどうかのフラグ

	// diaryEntitiesからselectedEntitiesを初期化
	// diaryEntitiesが変更されたら（ページロード時、または日記が再取得された時）
	// ただし、ユーザーが入力中（isTyping）またはエンティティ選択中（isSelectingEntity）の場合は更新しない
	$: {
		if (
			diaryEntities &&
			diaryEntities.length > 0 &&
			!isTyping &&
			!isSelectingEntity &&
			allEntities &&
			allEntities.length > 0
		) {
			// diaryEntitiesを検証してから使用
			// バックエンドから取得したデータに無効なエンティティが含まれている可能性があるため
			const validatedDiaryEntities = validateDiaryEntities(
				value,
				diaryEntities,
				allEntities,
			);

			// validatedDiaryEntitiesからselectedEntitiesを生成
			const entitiesFromDiary = validatedDiaryEntities
				.map((de) => ({
					entityId: de.entityId,
					positions: de.positions.map((pos) => ({
						start: pos.start,
						end: pos.end,
					})),
				}))
				.filter((e) => e.entityId !== "");

			// selectedEntitiesが空の場合は無条件で更新
			if (selectedEntities.length === 0) {
				selectedEntities = entitiesFromDiary;
			} else {
				// selectedEntitiesに既にデータがある場合、validatedDiaryEntitiesと比較
				// validatedDiaryEntitiesの方が新しいデータ（より多くのposition）を持っている場合のみ更新
				const selectedStr = JSON.stringify(selectedEntities);
				const diaryStr = JSON.stringify(entitiesFromDiary);

				// 完全一致しない場合のみチェック
				if (selectedStr !== diaryStr) {
					// selectedEntitiesの方がpositionが多い場合は、ユーザーが新しく追加した可能性があるので上書きしない
					const selectedTotalPositions = selectedEntities.reduce(
						(sum, e) => sum + e.positions.length,
						0,
					);
					const diaryTotalPositions = entitiesFromDiary.reduce(
						(sum, e) => sum + e.positions.length,
						0,
					);

					if (diaryTotalPositions > selectedTotalPositions) {
						selectedEntities = entitiesFromDiary;
					}
				}
			}
		}
		// diaryEntitiesが空の場合は何もしない
		// （ユーザーが新しくentityを選択している可能性があるため、クリアしない）
	}

	// diaryEntitiesが外部から変更されたら、保存されたコンテンツを更新
	$: if (diaryEntities !== previousDiaryEntities) {
		previousDiaryEntities = diaryEntities;
		savedContent = value;
	}

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

	// エンティティハイライトを適用したHTMLを生成
	// 入力中でない場合のみエンティティハイライトを表示
	// また、現在のvalueが保存されたコンテンツと一致する場合のみハイライトを適用
	$: highlightedHTML = (() => {
		// 入力中は常にプレーンテキスト
		if (isTyping) {
			return value.replace(/\n/g, "<br>");
		}

		// diaryEntitiesとdiaryHighlightsの両方がない場合はプレーンテキスト
		if (
			(!diaryEntities || diaryEntities.length === 0) &&
			(!diaryHighlights || diaryHighlights.length === 0)
		) {
			return value.replace(/\n/g, "<br>");
		}

		// 現在のvalueが保存されたコンテンツと異なる場合はプレーンテキスト
		// （編集中のテキストには古いpositionデータを適用しない）
		if (value !== savedContent) {
			return value.replace(/\n/g, "<br>");
		}

		// 保存されたコンテンツと一致する場合は、検証してからハイライトを適用
		// バックエンドから取得したdiaryEntitiesに無効なエンティティが含まれている可能性があるため
		const validatedEntities = validateDiaryEntities(
			value,
			diaryEntities,
			allEntities,
		);

		// diaryHighlightsがある場合は、エンティティとハイライトの両方を適用
		if (diaryHighlights && diaryHighlights.length > 0) {
			return highlightEntitiesAndHighlights(
				value,
				validatedEntities,
				diaryHighlights,
			);
		}

		return highlightEntities(value, validatedEntities);
	})();

	// captureフェーズでTabキーをキャプチャするためのリスナー
	onMount(async () => {
		// savedContentを確実に初期化（SSR時のリアクティビティの問題を回避）
		if (!savedContent) {
			savedContent = value;
		}

		// 全エンティティデータを事前取得
		await _loadAllEntities();

		// ブラウザ環境でのみイベントリスナーを追加
		if (typeof window !== "undefined") {
			// entity更新イベントをリッスン
			window.addEventListener("entityUpdated", handleEntityUpdated);
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
		// ブラウザ環境でのみイベントリスナーを削除
		if (typeof window !== "undefined") {
			// entity更新イベントリスナーを削除
			window.removeEventListener("entityUpdated", handleEntityUpdated);
		}
	});

	// entity更新イベントハンドラー
	function handleEntityUpdated() {
		// エンティティデータを再取得
		_loadAllEntities();
	}

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

	function _localRestoreCursorPosition(targetPos: number) {
		if (!contentElement) return;
		// SSR時は何もしない
		if (typeof window === "undefined") return;

		const selection = window.getSelection();
		if (!selection) return;

		let currentPos = 0;
		let targetNode: Node | null = null;
		let targetOffset = 0;
		let found = false;

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
				} else {
					// <a>タグなどの要素ノードの子を再帰的に処理
					for (const child of Array.from(node.childNodes)) {
						if (traverse(child)) return true;
					}
				}
			}
			return false;
		}

		found = traverse(contentElement);

		if (found && targetNode) {
			try {
				const range = document.createRange();
				// Node型として明示的にキャスト
				const node = targetNode as Node;
				// テキストノードの場合
				if (node.nodeType === Node.TEXT_NODE) {
					const textLength = node.textContent?.length || 0;
					range.setStart(node, Math.min(targetOffset, textLength));
					range.collapse(true);
				} else {
					// 要素ノードの場合（BRの親など）
					range.setStart(node, targetOffset);
					range.collapse(true);
				}
				selection.removeAllRanges();
				selection.addRange(range);
			} catch (e) {
				console.error("Failed to restore cursor position:", e);
				// カーソル復元に失敗した場合は末尾に配置
				try {
					const range = document.createRange();
					range.selectNodeContents(contentElement);
					range.collapse(false);
					selection.removeAllRanges();
					selection.addRange(range);
				} catch {
					// 何もしない
				}
			}
		} else {
			// targetPosが見つからない場合、createRangeAtTextOffsetを使って再試行
			try {
				const range = createRangeAtTextOffset(contentElement, targetPos);
				if (range) {
					selection.removeAllRanges();
					selection.addRange(range);
				} else {
					// それでも失敗した場合は末尾に配置
					const fallbackRange = document.createRange();
					fallbackRange.selectNodeContents(contentElement);
					fallbackRange.collapse(false);
					selection.removeAllRanges();
					selection.addRange(fallbackRange);
				}
			} catch {
				// 何もしない
			}
		}
	}

	async function _handleInput(event: Event) {
		const target = event.target as HTMLDivElement;
		isUpdatingFromValue = true;
		isTyping = true; // 入力中フラグを立てる

		// 既存のタイムアウトをクリア
		if (updateTimeout !== null) {
			clearTimeout(updateTimeout);
		}

		// IME入力中（compositionupdate）の場合は、エンティティ候補検索をスキップ
		// valueの更新のみ行う
		const isCompositionUpdate = event instanceof CompositionEvent;

		const oldValue = value;
		value = htmlToPlainText(target.innerHTML);

		// contentElementが初期化されていない場合は何もしない
		if (!contentElement) {
			isUpdatingFromValue = false;
			return;
		}

		// エンティティ内部編集の検出と紐づけ解除
		// DOM内の全てのリンク要素をチェックし、編集されたリンクを解除
		// その後、残っているリンクからselectedEntitiesを再構築
		const links = contentElement.querySelectorAll("a");
		const linksToRemove: HTMLAnchorElement[] = [];

		for (const link of links) {
			const href = link.getAttribute("href");
			if (!href?.includes("/entity/")) continue;

			const entityId = href.split("/entity/")[1];
			if (!entityId) continue;

			const linkText = link.textContent || "";

			// allEntitiesからこのentityIdのエンティティを取得
			const entity = allEntities.find((e) => e.id === entityId);
			if (!entity) continue; // エンティティが見つからない場合はスキップ

			// このエンティティの有効なテキスト（名前とエイリアス）を収集
			const validTexts = [entity.name];
			for (const alias of entity.aliases) {
				validTexts.push(alias.alias);
			}

			// linkTextが有効なテキストのいずれかと完全一致するかチェック
			const isValid = validTexts.some((text) => text === linkText);

			if (!isValid) {
				// リンクのテキストが元のエンティティ名/エイリアスと異なる場合は編集されている
				linksToRemove.push(link as HTMLAnchorElement);
			}
		}

		// 無効なリンクをDOMから削除
		if (linksToRemove.length > 0) {
			for (const link of linksToRemove) {
				// DOMから削除（テキストノードに置き換え）
				const textNode = document.createTextNode(link.textContent || "");
				link.parentNode?.replaceChild(textNode, link);
			}
		}

		// selectedEntitiesをDOMの実際の状態と同期
		// DOMを信頼できる唯一の情報源として扱い、そこから再構築する
		selectedEntities = syncSelectedEntitiesFromDOM(
			contentElement,
			getTextOffset,
		);

		// IME入力中は候補検索をスキップ（DOM操作を最小限にしてIMEの動作を安定させる）
		if (isCompositionUpdate) {
			isUpdatingFromValue = false;

			// 500ms後に入力が止まったら入力完了フラグを下ろす
			updateTimeout = setTimeout(() => {
				isTyping = false;
			}, 500);

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

		// カーソル前の最後の改行以降のテキストを取得
		const lastNewlineIndex = beforeCursor.lastIndexOf("\n");
		const textAfterLastNewline =
			lastNewlineIndex >= 0
				? beforeCursor.substring(lastNewlineIndex + 1)
				: beforeCursor;

		// エンティティ確定直後の場合、確定したエンティティ名の部分を除外
		let searchText = textAfterLastNewline;
		if (justSelectedEntity && lastSelectedEntityText) {
			// 確定したエンティティ名を除去
			// 例: "natori" を確定後 "natorina" となった場合、"na" のみを検索
			if (searchText.endsWith(lastSelectedEntityText)) {
				searchText = searchText.substring(
					0,
					searchText.length - lastSelectedEntityText.length,
				);
			} else if (searchText.startsWith(lastSelectedEntityText)) {
				searchText = searchText.substring(lastSelectedEntityText.length);
			}
			justSelectedEntity = false; // フラグをリセット
			lastSelectedEntityText = ""; // リセット
		}

		// searchTextが空の場合は候補を表示しない
		// （改行直後や、何も入力していない状態では候補を出さない）
		if (searchText.length === 0) {
			showSuggestions = false;
			currentQuery = "";
			currentTriggerPos = -1;
		} else {
			// カーソル位置から後方に向かって、全てのエンティティとマッチするかチェック
			// 最長一致を優先する
			let bestMatch: { word: string; startPos: number } | null = null;

			// 後方から2文字以上の部分文字列を試す
			for (let len = searchText.length; len >= 2; len--) {
				const substring = searchText.substring(searchText.length - len);

				// このsubstringで始まるエンティティがあるかチェック
				const hasMatch = allFlatEntities.some((flat) =>
					flat.text.toLowerCase().startsWith(substring.toLowerCase()),
				);

				if (hasMatch) {
					bestMatch = {
						word: substring,
						startPos: cursorPos - len,
					};
					break; // 最長一致が見つかったので終了
				}
			}

			if (bestMatch) {
				currentTriggerPos = bestMatch.startPos;
				currentQuery = bestMatch.word;
				await searchForSuggestions(bestMatch.word);
				if (!showSuggestions || flatSuggestions.length === 0) {
					showSuggestions = false;
					currentQuery = "";
					currentTriggerPos = -1;
				}
			} else {
				// エンティティ候補がない場合は候補を閉じる
				showSuggestions = false;
				currentQuery = "";
				currentTriggerPos = -1;
			}
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

	// 候補検索
	// 全エンティティデータを事前取得
	async function _loadAllEntities() {
		const result = await loadAllEntities();
		allEntities = result.entities;
		allFlatEntities = result.flatEntities;
	}

	// ブラウザ側でエンティティをフィルタリング
	async function searchForSuggestions(query: string) {
		try {
			flatSuggestions = filterEntitiesByPrefix(query, allFlatEntities);
			showSuggestions = flatSuggestions.length > 0;
			if (showSuggestions) {
				selectedSuggestionIndex = getBestMatchIndex(query, flatSuggestions);
			}
		} catch (err) {
			console.error("Failed to search entities:", err);
			flatSuggestions = [];
			showSuggestions = false;
		}
	}

	// 候補表示位置を更新
	function updateSuggestionPosition(_target: HTMLDivElement) {
		// SSR時は何もしない
		if (typeof window === "undefined") return;

		// 現在のカーソル位置を取得
		const selection = window.getSelection();
		if (!selection || selection.rangeCount === 0) return;

		// rangeをcloneして元のrangeを保護
		const range = selection.getRangeAt(0).cloneRange();
		const rect = range.getBoundingClientRect();

		// collapsed range（カーソル位置）でrectが有効でない場合のみ、span要素を使用
		if (rect.width === 0 && rect.height === 0) {
			// 一時的なspan要素を挿入して位置を取得
			const tempSpan = document.createElement("span");
			tempSpan.textContent = "\u200B"; // ゼロ幅スペース
			range.insertNode(tempSpan);

			const spanRect = tempSpan.getBoundingClientRect();

			// span要素を削除
			tempSpan.remove();

			// 元のカーソル位置を復元（cloneしたrangeなので元の選択範囲には影響なし）
			// ただし、insertNode後にselectionが変わるため、元のrangeを復元
			const originalRange = selection.getRangeAt(0);
			selection.removeAllRanges();
			selection.addRange(originalRange);

			// position: fixedを使用するため、ビューポート座標をそのまま使用（スクロール量は不要）
			// カーソルの下に表示（5px下）
			suggestionPosition = {
				top: spanRect.bottom + 5,
				left: spanRect.left,
			};
		} else {
			// rectが有効な場合はそのまま使用
			// position: fixedを使用するため、ビューポート座標をそのまま使用（スクロール量は不要）
			suggestionPosition = {
				top: rect.bottom + 5,
				left: rect.left,
			};
		}
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
					indexToSelect = getBestMatchIndex(currentQuery, flatSuggestions);
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
				// <br>の後ろに意味のあるコンテンツ（テキストやエンティティリンクなど）が存在するかチェック
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

			// savedContentを更新して、エンティティハイライト復元を防ぐ
			// （改行直後は新しいコンテンツとして扱う）
			savedContent = "";
		}
	}

	// 候補選択
	async function selectSuggestion(entity: Entity, selectedText?: string) {
		if (currentTriggerPos === -1) return;

		// エンティティ選択中フラグを立てる（reactive statementからの上書きを防ぐ）
		isSelectingEntity = true;

		// DOM上の既存エンティティリンクからselectedEntitiesを再構築
		// _handleInputを経由せずに直接呼ばれる場合、DOM上のリンクとselectedEntitiesが同期していないため
		const links = contentElement.querySelectorAll("a[href*='/entity/']");
		const reconstructedEntities: {
			entityId: string;
			positions: { start: number; end: number }[];
		}[] = [];

		for (const link of Array.from(links)) {
			const href = link.getAttribute("href");
			if (!href?.includes("/entity/")) continue;

			const entityId = href.split("/entity/")[1];
			if (!entityId) continue;

			// リンクのテキスト位置を取得
			const linkStartOffset = getTextOffset(contentElement, link, 0);
			const linkText = link.textContent || "";
			const linkEndOffset = linkStartOffset + linkText.length;

			// 既存のエンティティに追加、または新規作成
			const existing = reconstructedEntities.find(
				(e) => e.entityId === entityId,
			);
			if (existing) {
				existing.positions.push({ start: linkStartOffset, end: linkEndOffset });
			} else {
				reconstructedEntities.push({
					entityId,
					positions: [{ start: linkStartOffset, end: linkEndOffset }],
				});
			}

			// リンクをプレーンテキストに置き換え
			const textNode = document.createTextNode(linkText);
			link.parentNode?.replaceChild(textNode, link);
		}

		// reconstructedEntitiesは後で使用するため、ここでは変数に保持するだけ
		// selectedEntitiesの更新は、新しいエンティティのposition調整と追加が完了してから行う

		// contentElementの最新の内容からvalueを更新（この時点でリンクは全て削除済み）
		value = htmlToPlainText(contentElement.innerHTML);

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

		// 選択されたテキストを使ってcurrentTriggerPosを再計算
		// currentQueryは信頼できないため、エンティティ名/エイリアスから直接検索
		const textToInsert = selectedText || entity.name;
		const beforeCursor = value.substring(0, cursorPos);

		// カーソル位置から逆方向に検索して、最も近い一致を見つける
		// （複数同一entityがある場合、カーソル位置に最も近いものを選択）
		let triggerPos = -1;

		// カーソル直前から後方に検索し、最長一致を見つける
		for (let i = 1; i <= beforeCursor.length; i++) {
			const partialText = beforeCursor.substring(beforeCursor.length - i);
			if (textToInsert.startsWith(partialText)) {
				triggerPos = beforeCursor.length - i;
				// 最長一致が見つかるまで継続（より長い一致があれば上書き）
			}
		}

		if (triggerPos !== -1) {
			currentTriggerPos = triggerPos;
		}

		// 入力中の単語の開始位置から現在のカーソル位置までを削除して置き換え
		const beforeTrigger = value.substring(0, currentTriggerPos);
		const afterCursor = value.substring(cursorPos);

		// currentTriggerPosからcursorPosまでの文字列を選択されたテキストに置き換え
		value = `${beforeTrigger}${textToInsert}${afterCursor}`;

		// 置き換えによって変化した文字数を計算
		// 元の文字列(currentTriggerPos～cursorPos)の長さ
		const oldLength = cursorPos - currentTriggerPos;
		// 新しい文字列の長さ
		const newLength = textToInsert.length;
		// 差分（正の場合は挿入、負の場合は削除）
		const lengthDiff = newLength - oldLength;

		// 再構築したentitiesのpositionを調整
		// 挿入位置(currentTriggerPos)より後ろにあるpositionを全て調整
		const adjustedEntities = reconstructedEntities
			.map((e) => {
				const adjustedPositions = e.positions
					.map((pos) => {
						// 置き換え範囲: currentTriggerPos ～ cursorPos
						const replaceStart = currentTriggerPos;
						const replaceEnd = cursorPos;

						// 挿入位置より前のpositionはそのまま（置き換え範囲の前）
						if (pos.end <= replaceStart) {
							return pos;
						}
						// 置き換え範囲と完全に重複するpositionは除外
						// （置き換え対象のエンティティそのもの）
						if (pos.start >= replaceStart && pos.end <= replaceEnd) {
							return null; // 除外
						}
						// 置き換え範囲とpositionが部分的に重複する場合も除外
						// （テキストの一部が置き換えられる場合、そのエンティティは無効になる）
						if (
							(pos.start < replaceStart &&
								pos.end > replaceStart &&
								pos.end <= replaceEnd) ||
							(pos.start >= replaceStart &&
								pos.start < replaceEnd &&
								pos.end > replaceEnd)
						) {
							return null; // 除外
						}
						// 置き換え範囲より後ろのpositionは調整
						if (pos.start >= replaceEnd) {
							return {
								start: pos.start + lengthDiff,
								end: pos.end + lengthDiff,
							};
						}
						// その他（開始が挿入位置より前で、終了が置き換え範囲より後ろ）
						// このケースは通常起こらないが、念のため調整
						return {
							start: pos.start,
							end: pos.end + lengthDiff,
						};
					})
					.filter((pos): pos is { start: number; end: number } => pos !== null);

				return {
					...e,
					positions: adjustedPositions,
				};
			})
			.filter((e) => e.positions.length > 0);

		// 選択されたエンティティの位置情報を記録
		const newPosition = {
			start: currentTriggerPos,
			end: currentTriggerPos + textToInsert.length,
		};

		// 調整済みのadjustedEntitiesから同じentityIdのものを探す
		const existingEntityIndex = adjustedEntities.findIndex(
			(e) => e.entityId === entity.id,
		);

		if (existingEntityIndex >= 0) {
			// 既に存在する場合は、positionsに追加
			// Svelteのreactivityのために新しい配列を作成
			selectedEntities = adjustedEntities.map((e, idx) =>
				idx === existingEntityIndex
					? { ...e, positions: [...e.positions, newPosition] }
					: e,
			);
		} else {
			// 新しいentityの場合は追加
			selectedEntities = [
				...adjustedEntities,
				{
					entityId: entity.id,
					positions: [newPosition],
				},
			];
		}

		showSuggestions = false;
		selectedSuggestionIndex = -1;
		currentTriggerPos = -1;

		// エンティティ確定直後フラグを立てる
		justSelectedEntity = true;
		lastSelectedEntityText = textToInsert;

		// 既存のタイムアウトをクリア
		if (updateTimeout !== null) {
			clearTimeout(updateTimeout);
		}

		// 入力中フラグを下ろす
		isTyping = false;

		// IME入力中フラグもクリア（エンティティ確定時はIME確定と同等の扱い）
		// スマホでのIME入力中にエンティティを確定した場合、IMEの未確定文字が重複しないようにする
		isComposing = false;

		// 次のティック（reactive statements実行後）まで待つ
		await tick();

		// エンティティ選択完了フラグを下ろす
		// tick()後に実行することで、selectedEntitiesの更新後に reactive statement が実行される
		isSelectingEntity = false;

		// contentEditableの内容を更新してフォーカスを戻す
		// カーソル位置を設定（エンティティテキストの直後）
		const newCursorPos = beforeTrigger.length + textToInsert.length;

		setTimeout(() => {
			// selectedEntitiesからエンティティハイライトを適用したHTMLを生成
			const htmlWithEntities = generateEntityHighlightHTML(
				value,
				selectedEntities,
			);
			contentElement.innerHTML = htmlWithEntities;

			// カーソル位置を復元してからフォーカス
			// エンティティ確定後は必ずエンティティの直後にカーソルを配置する
			try {
				_localRestoreCursorPosition(newCursorPos);
			} catch (e) {
				console.error("Failed to restore cursor after entity selection:", e);
				// カーソル復元に失敗した場合は、エンティティの直後（newCursorPos）に配置を試みる
				try {
					const range = createRangeAtTextOffset(contentElement, newCursorPos);
					if (range) {
						const selection = window.getSelection();
						if (selection) {
							selection.removeAllRanges();
							selection.addRange(range);
						}
					}
				} catch {
					// それでも失敗した場合は末尾に配置
					const range = document.createRange();
					const selection = window.getSelection();
					if (contentElement.lastChild) {
						range.setStartAfter(contentElement.lastChild);
						range.collapse(true);
						selection?.removeAllRanges();
						selection?.addRange(range);
					}
				}
			}
			contentElement.focus();
		}, 0);
	}

	// Update content when value changes externally
	// ただし、updateContentElement()が呼ばれる条件の場合はスキップ
	// （エンティティハイライトが適用される場合は、updateContentElement()に任せる）
	// また、エンティティ選択中（isSelectingEntity）の場合もスキップ
	$: if (
		contentElement &&
		htmlToPlainText(contentElement.innerHTML) !== value &&
		!isSelectingEntity &&
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
		on:compositionupdate={_handleInput}
		on:compositionend={(event) => {
			isComposing = false;
			// IME確定後に候補検索を実行
			_handleInput(event);
		}}
		{...$$restProps}
	></div>
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
