/**
 * contenteditable要素のカーソル位置管理ユーティリティ
 */

/**
 * contenteditable要素内のテキストオフセット位置を取得
 * @param root ルート要素
 * @param node 対象ノード
 * @param offset ノード内のオフセット
 * @returns テキストオフセット位置
 */
export function getTextOffset(root: Node, node: Node, offset: number): number {
	let textOffset = 0;

	function traverse(currentNode: Node): number | null {
		if (currentNode === node) {
			if (currentNode.nodeType === Node.TEXT_NODE) {
				return textOffset + offset;
			}
			if (currentNode.nodeType === Node.ELEMENT_NODE) {
				const children = Array.from(currentNode.childNodes);
				for (let i = 0; i < Math.min(offset, children.length); i++) {
					const child = children[i];
					if (child.nodeType === Node.TEXT_NODE) {
						textOffset += child.textContent?.length || 0;
					} else if (child.nodeType === Node.ELEMENT_NODE) {
						if (child.nodeName === "BR") {
							textOffset += 1;
						} else {
							textOffset += getTextLength(child);
						}
					}
				}
				return textOffset;
			}
		}

		if (currentNode.nodeType === Node.TEXT_NODE) {
			textOffset += currentNode.textContent?.length || 0;
		} else if (currentNode.nodeType === Node.ELEMENT_NODE) {
			if (currentNode.nodeName === "BR") {
				textOffset += 1;
			}
			for (const child of Array.from(currentNode.childNodes)) {
				const result = traverse(child);
				if (result !== null) return result;
			}
		}

		return null;
	}

	function getTextLength(node: Node): number {
		if (node.nodeType === Node.TEXT_NODE) {
			return node.textContent?.length || 0;
		}
		if (node.nodeType === Node.ELEMENT_NODE) {
			if (node.nodeName === "BR") {
				return 1;
			}
			let length = 0;
			for (const child of Array.from(node.childNodes)) {
				length += getTextLength(child);
			}
			return length;
		}
		return 0;
	}

	const result = traverse(root);
	return result !== null ? result : textOffset;
}

/**
 * テキストオフセット位置にRangeを作成
 * @param root ルート要素
 * @param targetOffset 目標テキストオフセット
 * @returns 作成されたRange、または失敗時null
 */
export function createRangeAtTextOffset(
	root: Node,
	targetOffset: number,
): Range | null {
	const range = document.createRange();
	let currentOffset = 0;

	function traverse(currentNode: Node): boolean {
		if (currentNode.nodeType === Node.TEXT_NODE) {
			const textLength = currentNode.textContent?.length || 0;
			if (currentOffset + textLength >= targetOffset) {
				const offset = targetOffset - currentOffset;
				range.setStart(currentNode, Math.min(offset, textLength));
				range.collapse(true);
				return true;
			}
			currentOffset += textLength;
		} else if (currentNode.nodeType === Node.ELEMENT_NODE) {
			if (currentNode.nodeName === "BR") {
				if (currentOffset === targetOffset) {
					const parent = currentNode.parentNode;
					if (parent) {
						const index = Array.from(parent.childNodes).indexOf(
							currentNode as ChildNode,
						);
						range.setStart(parent, index);
						range.collapse(true);
						return true;
					}
				}
				currentOffset += 1;
			}
			for (const child of Array.from(currentNode.childNodes)) {
				if (traverse(child)) return true;
			}
		}

		return false;
	}

	if (traverse(root)) {
		return range;
	}

	// オフセットが範囲外の場合は最後に設定
	if (root.lastChild) {
		range.setStartAfter(root.lastChild);
		range.collapse(true);
		return range;
	}

	return null;
}

/**
 * カーソル位置を復元
 * @param contentElement contenteditable要素
 * @param targetPos 目標テキストオフセット
 */
export function restoreCursorPosition(
	contentElement: HTMLDivElement,
	targetPos: number,
): void {
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
			const node = targetNode as Node;
			if (node.nodeType === Node.TEXT_NODE) {
				const textLength = node.textContent?.length || 0;
				range.setStart(node, Math.min(targetOffset, textLength));
				range.collapse(true);
			} else {
				range.setStart(node, targetOffset);
				range.collapse(true);
			}
			selection.removeAllRanges();
			selection.addRange(range);
		} catch (e) {
			console.error("Failed to restore cursor position:", e);
			fallbackToEnd(contentElement, selection);
		}
	} else {
		// createRangeAtTextOffsetを使って再試行
		try {
			const range = createRangeAtTextOffset(contentElement, targetPos);
			if (range) {
				selection.removeAllRanges();
				selection.addRange(range);
			} else {
				fallbackToEnd(contentElement, selection);
			}
		} catch {
			fallbackToEnd(contentElement, selection);
		}
	}
}

/**
 * カーソルを末尾に配置（フォールバック）
 */
function fallbackToEnd(contentElement: HTMLDivElement, selection: Selection) {
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

/**
 * 現在のカーソル位置を保存
 * @returns 保存されたRange、または失敗時null
 */
export function saveCursorPosition(): Range | null {
	const selection = window.getSelection();
	if (selection && selection.rangeCount > 0) {
		return selection.getRangeAt(0);
	}
	return null;
}

/**
 * Rangeからカーソル位置を復元
 * @param range 保存されたRange
 */
export function restoreCursorFromRange(range: Range): void {
	const selection = window.getSelection();
	if (selection && range) {
		selection.removeAllRanges();
		selection.addRange(range);
	}
}
