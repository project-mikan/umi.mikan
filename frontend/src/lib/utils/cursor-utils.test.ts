/**
 * cursor-utilsのテスト
 */
import { describe, expect, it, beforeEach } from "vitest";
import {
	getTextOffset,
	createRangeAtTextOffset,
	restoreCursorPosition,
	saveCursorPosition,
	restoreCursorFromRange,
} from "./cursor-utils";

describe("cursor-utils", () => {
	let container: HTMLDivElement;

	beforeEach(() => {
		// contenteditable要素を作成
		container = document.createElement("div");
		container.contentEditable = "true";
		document.body.appendChild(container);
	});

	describe("getTextOffset", () => {
		it("単純なテキストノードのオフセットを正しく計算する", () => {
			container.textContent = "Hello World";
			const textNode = container.firstChild as Node;
			expect(getTextOffset(container, textNode, 0)).toBe(0);
			expect(getTextOffset(container, textNode, 5)).toBe(5);
			expect(getTextOffset(container, textNode, 11)).toBe(11);
		});

		it("BRタグを含むテキストのオフセットを正しく計算する", () => {
			container.innerHTML = "Hello<br>World";
			const firstTextNode = container.firstChild as Node;
			const secondTextNode = container.childNodes[2] as Node;

			expect(getTextOffset(container, firstTextNode, 5)).toBe(5);
			// BRタグの位置（5文字 + 1文字（BR））
			expect(getTextOffset(container, secondTextNode, 0)).toBe(6);
			expect(getTextOffset(container, secondTextNode, 5)).toBe(11);
		});

		it("リンク要素を含むテキストのオフセットを正しく計算する", () => {
			container.innerHTML = 'Text <a href="#">Link</a> More';
			const firstTextNode = container.firstChild as Node;
			const linkNode = container.childNodes[1] as Node;
			const linkTextNode = linkNode.firstChild as Node;
			const lastTextNode = container.childNodes[2] as Node;

			expect(getTextOffset(container, firstTextNode, 5)).toBe(5);
			expect(getTextOffset(container, linkTextNode, 0)).toBe(5);
			expect(getTextOffset(container, linkTextNode, 4)).toBe(9);
			expect(getTextOffset(container, lastTextNode, 0)).toBe(9);
		});

		it("ネストされた要素のオフセットを正しく計算する", () => {
			container.innerHTML = "Start<div>Nested<br>Text</div>End";
			const startTextNode = container.firstChild as Node;
			const divNode = container.childNodes[1] as Node;
			const nestedTextNode = divNode.firstChild as Node;
			const textAfterBr = divNode.childNodes[2] as Node;
			const endTextNode = container.childNodes[2] as Node;

			expect(getTextOffset(container, startTextNode, 5)).toBe(5);
			expect(getTextOffset(container, nestedTextNode, 6)).toBe(11);
			// BR後（5 + 6 + 1（BR））
			expect(getTextOffset(container, textAfterBr, 0)).toBe(12);
			expect(getTextOffset(container, textAfterBr, 4)).toBe(16);
			expect(getTextOffset(container, endTextNode, 0)).toBe(16);
		});

		it("空のコンテナで0を返す", () => {
			container.innerHTML = "";
			expect(getTextOffset(container, container, 0)).toBe(0);
		});
	});

	describe("createRangeAtTextOffset", () => {
		it("単純なテキストノード内の位置にRangeを作成する", () => {
			container.textContent = "Hello World";
			const range = createRangeAtTextOffset(container, 6);

			expect(range).not.toBeNull();
			if (range) {
				expect(range.startContainer).toBe(container.firstChild);
				expect(range.startOffset).toBe(6);
				expect(range.collapsed).toBe(true);
			}
		});

		it("BRタグの前後の位置にRangeを作成する", () => {
			container.innerHTML = "Hello<br>World";

			// BRの直前
			const range1 = createRangeAtTextOffset(container, 5);
			expect(range1).not.toBeNull();

			// BRの直後
			const range2 = createRangeAtTextOffset(container, 6);
			expect(range2).not.toBeNull();
			if (range2) {
				expect(range2.collapsed).toBe(true);
			}
		});

		it("リンク要素内の位置にRangeを作成する", () => {
			container.innerHTML = 'Text <a href="#">Link</a> More';
			const range = createRangeAtTextOffset(container, 7); // "Link"の2文字目

			expect(range).not.toBeNull();
			if (range) {
				expect(range.collapsed).toBe(true);
			}
		});

		it("範囲外のオフセットで最後にRangeを作成する", () => {
			container.textContent = "Hello";
			const range = createRangeAtTextOffset(container, 100);

			expect(range).not.toBeNull();
			if (range) {
				expect(range.collapsed).toBe(true);
			}
		});

		it("空のコンテナでRangeを作成する", () => {
			container.innerHTML = "";
			const range = createRangeAtTextOffset(container, 0);

			// 空の場合はnullが返される
			expect(range).toBeNull();
		});
	});

	describe("restoreCursorPosition", () => {
		it("指定した位置にカーソルを復元する", () => {
			container.textContent = "Hello World";

			restoreCursorPosition(container, 6);

			const selection = window.getSelection();
			expect(selection).not.toBeNull();
			if (selection && selection.rangeCount > 0) {
				const range = selection.getRangeAt(0);
				const offset = getTextOffset(
					container,
					range.startContainer,
					range.startOffset,
				);
				expect(offset).toBe(6);
			}
		});

		it("BRタグを含むテキストでカーソルを復元する", () => {
			container.innerHTML = "Hello<br>World";

			// BRの後（"World"の先頭）
			restoreCursorPosition(container, 6);

			const selection = window.getSelection();
			expect(selection).not.toBeNull();
			if (selection && selection.rangeCount > 0) {
				const range = selection.getRangeAt(0);
				expect(range.collapsed).toBe(true);
			}
		});

		it("リンク要素を含むテキストでカーソルを復元する", () => {
			container.innerHTML = 'Text <a href="#">Link</a> More';

			restoreCursorPosition(container, 7);

			const selection = window.getSelection();
			expect(selection).not.toBeNull();
			if (selection && selection.rangeCount > 0) {
				const range = selection.getRangeAt(0);
				expect(range.collapsed).toBe(true);
			}
		});

		it("範囲外のオフセットで末尾にカーソルを復元する", () => {
			container.textContent = "Hello";

			restoreCursorPosition(container, 100);

			const selection = window.getSelection();
			expect(selection).not.toBeNull();
			if (selection && selection.rangeCount > 0) {
				const range = selection.getRangeAt(0);
				expect(range.collapsed).toBe(true);
			}
		});
	});

	describe("saveCursorPosition", () => {
		it("現在のカーソル位置を保存する", () => {
			container.textContent = "Hello World";

			// カーソル位置を設定
			const selection = window.getSelection();
			if (selection) {
				const range = document.createRange();
				const textNode = container.firstChild as Node;
				range.setStart(textNode, 6);
				range.collapse(true);
				selection.removeAllRanges();
				selection.addRange(range);

				// 保存
				const savedRange = saveCursorPosition();
				expect(savedRange).not.toBeNull();
				if (savedRange) {
					expect(savedRange.startContainer).toBe(textNode);
					expect(savedRange.startOffset).toBe(6);
				}
			}
		});

		it("選択範囲がない場合nullを返す", () => {
			const selection = window.getSelection();
			if (selection) {
				selection.removeAllRanges();
			}

			const savedRange = saveCursorPosition();
			expect(savedRange).toBeNull();
		});
	});

	describe("restoreCursorFromRange", () => {
		it("保存したRangeからカーソル位置を復元する", () => {
			container.textContent = "Hello World";

			// Rangeを作成して保存
			const range = document.createRange();
			const textNode = container.firstChild as Node;
			range.setStart(textNode, 6);
			range.collapse(true);

			// 復元
			restoreCursorFromRange(range);

			const selection = window.getSelection();
			expect(selection).not.toBeNull();
			if (selection && selection.rangeCount > 0) {
				const currentRange = selection.getRangeAt(0);
				expect(currentRange.startContainer).toBe(textNode);
				expect(currentRange.startOffset).toBe(6);
			}
		});
	});

	describe("統合テスト", () => {
		it("save -> restore のワークフローが正しく動作する", () => {
			container.textContent = "Hello World";

			// カーソル位置を設定
			restoreCursorPosition(container, 6);

			// 保存
			const savedRange = saveCursorPosition();
			expect(savedRange).not.toBeNull();

			// カーソルをクリア
			const selection = window.getSelection();
			if (selection) {
				selection.removeAllRanges();
			}

			// 復元
			if (savedRange) {
				restoreCursorFromRange(savedRange);
			}

			// 検証
			if (selection && selection.rangeCount > 0) {
				const range = selection.getRangeAt(0);
				const offset = getTextOffset(
					container,
					range.startContainer,
					range.startOffset,
				);
				expect(offset).toBe(6);
			}
		});

		it("複雑なHTML構造でsave -> restoreが正しく動作する", () => {
			container.innerHTML = 'Start<div>Middle<a href="#">Link</a></div><br>End';

			// "Link"の中にカーソルを設定
			restoreCursorPosition(container, 13);

			const savedRange = saveCursorPosition();
			expect(savedRange).not.toBeNull();

			// カーソルをクリア
			const selection = window.getSelection();
			if (selection) {
				selection.removeAllRanges();
			}

			// 復元
			if (savedRange) {
				restoreCursorFromRange(savedRange);
			}

			// 検証
			if (selection && selection.rangeCount > 0) {
				const range = selection.getRangeAt(0);
				expect(range.collapsed).toBe(true);
			}
		});
	});
});
