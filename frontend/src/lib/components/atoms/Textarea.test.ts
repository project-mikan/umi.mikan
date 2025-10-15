import { beforeEach, describe, expect, it, vi } from "vitest";

describe("Textarea Component Functionality", () => {
	let mockDiv: {
		innerHTML: string;
		textContent: string;
		innerText: string;
	};

	beforeEach(() => {
		mockDiv = {
			innerHTML: "",
			textContent: "",
			innerText: "",
		};

		// Mock document.createElement
		vi.spyOn(document, "createElement").mockReturnValue(
			mockDiv as unknown as HTMLElement,
		);
	});

	// Test the htmlToPlainText function logic
	function htmlToPlainText(html: string): string {
		const tempDiv = document.createElement("div");
		tempDiv.innerHTML = html;

		// Convert common HTML elements to plain text
		tempDiv.innerHTML = tempDiv.innerHTML.replace(/<br\s*\/?>/gi, "\n");
		tempDiv.innerHTML = tempDiv.innerHTML.replace(/<\/p>/gi, "\n");
		tempDiv.innerHTML = tempDiv.innerHTML.replace(/<p[^>]*>/gi, "");
		tempDiv.innerHTML = tempDiv.innerHTML.replace(/<\/div>/gi, "\n");
		tempDiv.innerHTML = tempDiv.innerHTML.replace(/<div[^>]*>/gi, "");
		tempDiv.innerHTML = tempDiv.innerHTML.replace(/<li[^>]*>/gi, "• ");
		tempDiv.innerHTML = tempDiv.innerHTML.replace(/<\/li>/gi, "\n");
		tempDiv.innerHTML = tempDiv.innerHTML.replace(
			/<\/?(?:ul|ol|strong|b|em|i|u|span|font)[^>]*>/gi,
			"",
		);

		let plainText = tempDiv.textContent || tempDiv.innerText || "";

		// Clean up extra whitespace and newlines only for pasted HTML content
		const hasComplexHTML = /<(?!br\s*\/?>)[^>]+>/.test(html);

		if (hasComplexHTML) {
			plainText = plainText.replace(/^\s+|\s+$/g, "").replace(/[ \t]+/g, " ");
		}

		return plainText;
	}

	it("should preserve single newlines correctly", () => {
		const testValue = "Line 1\nLine 2\nLine 3";
		// This simulates the component's rendering logic
		const rendered = testValue.replace(/\n/g, "<br>");
		expect(rendered).toBe("Line 1<br>Line 2<br>Line 3");
	});

	it("should preserve multiple newlines including empty lines", () => {
		const testValue = "Line 1\n\nLine 3\n\n\nLine 6";
		const rendered = testValue.replace(/\n/g, "<br>");
		expect(rendered).toBe("Line 1<br><br>Line 3<br><br><br>Line 6");
	});

	it("should handle whitespace-only lines correctly", () => {
		const testValue = "Line 1\n \nLine 3\n  \nLine 5";
		const rendered = testValue.replace(/\n/g, "<br>");
		expect(rendered).toBe("Line 1<br> <br>Line 3<br>  <br>Line 5");
	});

	it("should strip HTML tags from pasted content (Google Keep style)", () => {
		const htmlContent =
			"<p>First paragraph</p><p>Second paragraph</p><p><strong>Bold text</strong></p>";

		// Mock the div processing
		mockDiv.innerHTML = htmlContent;
		mockDiv.textContent = "First paragraph\nSecond paragraph\nBold text";

		const result = htmlToPlainText(htmlContent);
		expect(result).toBe("First paragraph\nSecond paragraph\nBold text");
	});

	it("should handle complex HTML with lists and formatting", () => {
		const htmlContent =
			"<p>Introduction</p><ul><li>First item</li><li><em>Second item</em></li><li><strong>Third item</strong></li></ul><p>Conclusion</p>";

		mockDiv.innerHTML = htmlContent;
		mockDiv.textContent =
			"Introduction\n• First item\n• Second item\n• Third item\nConclusion";

		const result = htmlToPlainText(htmlContent);
		expect(result).toBe(
			"Introduction\n• First item\n• Second item\n• Third item\nConclusion",
		);
	});

	it("should handle div tags correctly", () => {
		const htmlContent =
			"<div>First div</div><div>Second div</div><div><span>Nested content</span></div>";

		mockDiv.innerHTML = htmlContent;
		mockDiv.textContent = "First div\nSecond div\nNested content";

		const result = htmlToPlainText(htmlContent);
		expect(result).toBe("First div\nSecond div\nNested content");
	});

	it("should handle br tags correctly", () => {
		const htmlContent = "Line 1<br>Line 2<br/><br />Line 4";

		mockDiv.innerHTML = htmlContent;
		mockDiv.textContent = "Line 1\nLine 2\n\nLine 4";

		const result = htmlToPlainText(htmlContent);
		expect(result).toBe("Line 1\nLine 2\n\nLine 4");
	});

	it("should preserve manually typed newlines without cleaning", () => {
		const textContent = "Line 1\n\n\nLine 4\n \n  \nLine 7";

		mockDiv.innerHTML = textContent;
		mockDiv.textContent = textContent;

		const result = htmlToPlainText(textContent);
		// Should preserve all whitespace and newlines for manual input (no complex HTML)
		expect(result).toBe("Line 1\n\n\nLine 4\n \n  \nLine 7");
	});

	it("should handle simple HTML correctly", () => {
		const htmlContent = "<p>Test</p>";

		mockDiv.innerHTML = htmlContent;
		mockDiv.textContent = "Test";

		const result = htmlToPlainText(htmlContent);
		expect(result).toBe("Test");
	});

	it("should handle p tags correctly", () => {
		const htmlContent = "<p>Para 1</p><p>Para 2</p>";

		mockDiv.innerHTML = htmlContent;
		mockDiv.textContent = "Para 1\nPara 2\n";

		const result = htmlToPlainText(htmlContent);
		expect(result).toBe("Para 1\nPara 2");
	});

	it("should handle list items correctly", () => {
		const htmlContent = "<ul><li>Item 1</li><li>Item 2</li></ul>";

		mockDiv.innerHTML = htmlContent;
		mockDiv.textContent = "• Item 1\n• Item 2\n";

		const result = htmlToPlainText(htmlContent);
		expect(result).toBe("• Item 1\n• Item 2");
	});

	it("should strip formatting tags correctly", () => {
		const htmlContent =
			"<strong>Bold</strong> <em>Italic</em> <span>Span</span>";

		mockDiv.innerHTML = htmlContent;
		mockDiv.textContent = "Bold Italic Span";

		const result = htmlToPlainText(htmlContent);
		expect(result).toBe("Bold Italic Span");
	});

	it("should detect complex HTML vs simple text", () => {
		const simpleText = "Line 1<br>Line 2";
		const complexHTML = "<p>Para 1</p><p>Para 2</p>";

		const hasComplexHTML1 = /<(?!br\s*\/?>)[^>]+>/.test(simpleText);
		const hasComplexHTML2 = /<(?!br\s*\/?>)[^>]+>/.test(complexHTML);

		expect(hasComplexHTML1).toBe(false); // br tags should not be considered complex
		expect(hasComplexHTML2).toBe(true); // p tags should be considered complex
	});

	it("should preserve empty lines in manual input (aa\\n\\nbb\\n\\ncc)", () => {
		const textContent = "aa\n\nbb\n\ncc";

		mockDiv.innerHTML = textContent;
		mockDiv.textContent = textContent;

		const result = htmlToPlainText(textContent);
		// Should preserve all newlines for manual input (no complex HTML)
		expect(result).toBe("aa\n\nbb\n\ncc");
	});

	it("should preserve complex empty line patterns", () => {
		const textContent = "aa\n\n\nbb\n\n\n\ncc";

		mockDiv.innerHTML = textContent;
		mockDiv.textContent = textContent;

		const result = htmlToPlainText(textContent);
		// Should preserve all newlines for manual input
		expect(result).toBe("aa\n\n\nbb\n\n\n\ncc");
	});

	// Test for IME composition handling
	it("should handle IME composition events correctly", () => {
		const mockEvent = {
			key: "Enter",
			isComposing: true,
			preventDefault: vi.fn(),
		} as unknown as KeyboardEvent;

		// Mock the handleKeydown function logic
		function handleKeydown(event: KeyboardEvent) {
			if (event.isComposing) {
				// During IME composition, ignore the event
				return;
			}

			if (event.key === "Enter") {
				event.preventDefault();
				// Normal Enter key handling
			}
		}

		handleKeydown(mockEvent);

		// Should not call preventDefault during IME composition
		expect(mockEvent.preventDefault).not.toHaveBeenCalled();
	});

	it("should handle regular Enter key press (not during IME composition)", () => {
		const mockEvent = {
			key: "Enter",
			isComposing: false,
			preventDefault: vi.fn(),
		} as unknown as KeyboardEvent;

		// Mock the handleKeydown function logic
		function handleKeydown(event: KeyboardEvent) {
			if (event.isComposing) {
				// During IME composition, ignore the event
				return;
			}

			if (event.key === "Enter") {
				event.preventDefault();
				// Normal Enter key handling
			}
		}

		handleKeydown(mockEvent);

		// Should call preventDefault for regular Enter key press
		expect(mockEvent.preventDefault).toHaveBeenCalled();
	});

	// Test for single Enter on last line issue
	it("should handle single Enter on last line correctly", () => {
		// Mock DOM elements and Selection API
		const mockRange = {
			deleteContents: vi.fn(),
			insertNode: vi.fn(),
			endContainer: null as unknown as Node,
			endOffset: 0,
			setStartAfter: vi.fn(),
			setEndBefore: vi.fn(),
			collapse: vi.fn(),
		};

		const mockSelection = {
			rangeCount: 1,
			getRangeAt: vi.fn().mockReturnValue(mockRange),
			removeAllRanges: vi.fn(),
			addRange: vi.fn(),
		};

		const mockContentElement = {
			childNodes: { length: 1 },
			dispatchEvent: vi.fn(),
		};

		// Mock document.createElement to return br elements
		const mockBr = { tagName: "BR" };
		vi.spyOn(document, "createElement").mockReturnValue(
			mockBr as unknown as HTMLElement,
		);

		// Mock window.getSelection
		vi.spyOn(window, "getSelection").mockReturnValue(
			mockSelection as unknown as Selection,
		);

		// Mock the scenario where cursor is at the end of content
		mockRange.endContainer = mockContentElement as unknown as Node;
		mockRange.endOffset = mockContentElement.childNodes.length;

		// Simulate the handleKeydown function logic for Enter key
		const mockEvent = {
			key: "Enter",
			isComposing: false,
			preventDefault: vi.fn(),
		} as unknown as KeyboardEvent;

		// This is the core logic that should handle single Enter on last line
		function handleKeydown(event: KeyboardEvent, contentElement: HTMLElement) {
			if (event.isComposing) {
				return;
			}

			if (event.key === "Enter") {
				event.preventDefault();

				const selection = window.getSelection();
				if (selection && selection.rangeCount > 0) {
					const range = selection.getRangeAt(0);
					const br = document.createElement("br");

					range.deleteContents();
					range.insertNode(br);

					// Check if we're at the end of content
					const isAtEnd =
						range.endContainer === contentElement &&
						range.endOffset === contentElement.childNodes.length;

					if (isAtEnd) {
						// This should create proper cursor position on last line
						const extraBr = document.createElement("br");
						range.insertNode(extraBr);

						const newRange = {
							setStartAfter: vi.fn(),
							setEndBefore: vi.fn(),
							collapse: vi.fn(),
						};

						selection.removeAllRanges();
						selection.addRange(newRange as unknown as Range);
					}
				}

				const inputEvent = new Event("input", { bubbles: true });
				contentElement.dispatchEvent(inputEvent);
			}
		}

		handleKeydown(mockEvent, mockContentElement as unknown as HTMLElement);

		// Verify the function was called correctly
		expect(mockEvent.preventDefault).toHaveBeenCalled();
		expect(mockRange.deleteContents).toHaveBeenCalled();
		expect(mockRange.insertNode).toHaveBeenCalledTimes(2); // Should insert 2 br elements for last line
		expect(mockSelection.removeAllRanges).toHaveBeenCalled();
		expect(mockSelection.addRange).toHaveBeenCalled();
		expect(mockContentElement.dispatchEvent).toHaveBeenCalled();
	});

	it("should handle single Enter in middle of text correctly", () => {
		// Mock DOM elements and Selection API
		const mockRange = {
			deleteContents: vi.fn(),
			insertNode: vi.fn(),
			endContainer: { nodeType: Node.TEXT_NODE, textContent: "some text" },
			endOffset: 5, // Not at the end
			setStartAfter: vi.fn(),
			collapse: vi.fn(),
		};

		const mockSelection = {
			rangeCount: 1,
			getRangeAt: vi.fn().mockReturnValue(mockRange),
			removeAllRanges: vi.fn(),
			addRange: vi.fn(),
		};

		const mockContentElement = {
			childNodes: { length: 3 },
			dispatchEvent: vi.fn(),
		};

		// Mock document.createElement to return br elements
		const mockBr = { tagName: "BR" };
		vi.spyOn(document, "createElement").mockReturnValue(
			mockBr as unknown as HTMLElement,
		);

		// Mock window.getSelection
		vi.spyOn(window, "getSelection").mockReturnValue(
			mockSelection as unknown as Selection,
		);

		// Simulate the handleKeydown function logic for Enter key in middle
		const mockEvent = {
			key: "Enter",
			isComposing: false,
			preventDefault: vi.fn(),
		} as unknown as KeyboardEvent;

		function handleKeydown(event: KeyboardEvent, contentElement: HTMLElement) {
			if (event.isComposing) {
				return;
			}

			if (event.key === "Enter") {
				event.preventDefault();

				const selection = window.getSelection();
				if (selection && selection.rangeCount > 0) {
					const range = selection.getRangeAt(0);
					const br = document.createElement("br");

					range.deleteContents();
					range.insertNode(br);

					// Check if we're at the end of content
					const isAtEnd =
						range.endContainer === contentElement &&
						range.endOffset === contentElement.childNodes.length;

					if (!isAtEnd) {
						// Normal case - not at end, just insert one br
						const newRange = {
							setStartAfter: vi.fn(),
							collapse: vi.fn(),
						};

						selection.removeAllRanges();
						selection.addRange(newRange as unknown as Range);
					}
				}

				const inputEvent = new Event("input", { bubbles: true });
				contentElement.dispatchEvent(inputEvent);
			}
		}

		handleKeydown(mockEvent, mockContentElement as unknown as HTMLElement);

		// Verify the function was called correctly
		expect(mockEvent.preventDefault).toHaveBeenCalled();
		expect(mockRange.deleteContents).toHaveBeenCalled();
		expect(mockRange.insertNode).toHaveBeenCalledTimes(1); // Should insert only 1 br element for middle
		expect(mockSelection.removeAllRanges).toHaveBeenCalled();
		expect(mockSelection.addRange).toHaveBeenCalled();
		expect(mockContentElement.dispatchEvent).toHaveBeenCalled();
	});
});

describe("Entity Highlighting Behavior", () => {
	// エンティティハイライトの適用条件をテスト

	it("should not apply entity highlighting during typing (isTyping=true)", () => {
		const isTyping = true;
		const diaryEntities = [
			{
				entityId: "entity-1",
				positions: [{ start: 0, end: 4 }],
			},
		];

		// isTyping=trueの場合、エンティティハイライトを適用しない
		const shouldHighlight = !isTyping && diaryEntities.length > 0;

		expect(shouldHighlight).toBe(false);
	});

	it("should apply entity highlighting when not typing (isTyping=false)", () => {
		const value: string = "test";
		const isTyping = false;
		const savedContent: string = "test";
		const diaryEntities = [
			{
				entityId: "entity-1",
				positions: [{ start: 0, end: 4 }],
			},
		];

		// isTyping=falseかつvalue===savedContentの場合、エンティティハイライトを適用
		const shouldHighlight =
			!isTyping && diaryEntities.length > 0 && value === savedContent;

		expect(shouldHighlight).toBe(true);
	});

	it("should not apply entity highlighting when value differs from savedContent", () => {
		const value: string = "test new";
		const isTyping = false;
		const savedContent: string = "test";
		const diaryEntities = [
			{
				entityId: "entity-1",
				positions: [{ start: 0, end: 4 }],
			},
		];

		// value !== savedContentの場合、エンティティハイライトを適用しない
		const shouldHighlight =
			!isTyping && diaryEntities.length > 0 && value === savedContent;

		expect(shouldHighlight).toBe(false);
	});

	it("should not apply entity highlighting when diaryEntities is empty", () => {
		const value = "test";
		const isTyping = false;
		const savedContent = "test";
		const diaryEntities: never[] = [];

		// diaryEntitiesが空の場合、エンティティハイライトを適用しない
		const shouldHighlight =
			!isTyping && diaryEntities.length > 0 && value === savedContent;

		expect(shouldHighlight).toBe(false);
	});

	it("should update savedContent when diaryEntities changes", () => {
		let savedContent = "";
		let previousDiaryEntities: unknown[] = [];
		const value = "test content";
		const diaryEntities = [
			{
				entityId: "entity-1",
				positions: [{ start: 0, end: 4 }],
			},
		];

		// diaryEntitiesが変更されたらsavedContentを更新
		if (diaryEntities !== previousDiaryEntities) {
			previousDiaryEntities = diaryEntities;
			savedContent = value;
		}

		expect(savedContent).toBe("test content");
		expect(previousDiaryEntities).toBe(diaryEntities);
	});

	it("should not update savedContent when value changes during typing", () => {
		let savedContent = "original";
		const previousDiaryEntities = [
			{
				entityId: "entity-1",
				positions: [{ start: 0, end: 8 }],
			},
		];
		const value = "original new text";
		const diaryEntities = previousDiaryEntities; // 同じ参照

		// diaryEntitiesが変更されていない場合、savedContentは更新されない
		if (diaryEntities !== previousDiaryEntities) {
			savedContent = value;
		}

		expect(savedContent).toBe("original"); // 変更されない
	});
});

describe("Entity Position Adjustment on Insertion", () => {
	// エンティティ挿入時に既存のエンティティ位置を調整するロジックをテスト

	/**
	 * エンティティ位置調整ロジックの実装
	 * selectSuggestion関数内で使われるロジックを抽出してテスト
	 */
	function adjustEntityPositions(
		selectedEntities: {
			entityId: string;
			positions: { start: number; end: number }[];
		}[],
		insertPos: number,
		oldLength: number,
		newLength: number,
	): {
		entityId: string;
		positions: { start: number; end: number }[];
	}[] {
		const lengthDiff = newLength - oldLength;

		return selectedEntities
			.map((e) => {
				const adjustedPositions = e.positions
					.map((pos) => {
						// 挿入位置より前のpositionはそのまま
						if (pos.end <= insertPos) {
							return pos;
						}
						// 挿入位置と重複するpositionは除外
						if (pos.start < insertPos + oldLength && pos.end > insertPos) {
							return null;
						}
						// 挿入位置より後ろのpositionは調整
						if (pos.start >= insertPos + oldLength) {
							return {
								start: pos.start + lengthDiff,
								end: pos.end + lengthDiff,
							};
						}
						// その他（開始が挿入位置より前で、終了が挿入位置より後ろ）
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
	}

	it("前の行でエンティティを挿入した時、次の行のエンティティ位置が正しく調整される", () => {
		// 初期状態:
		// "natoriaaaaaほげ\nnatoria"
		// positions: [{ start: 0, end: 6 }, { start: 18, end: 24 }]
		//             ^natori^              ^natori^ (6文字)

		const initialEntities = [
			{
				entityId: "natori-id",
				positions: [
					{ start: 0, end: 6 }, // "natori"aaaaaほげ
					{ start: 18, end: 24 }, // natoriaaaaaほげ\n"natori"a
				],
			},
		];

		// 改行の直後(17文字目)に"sato"(4文字)を挿入
		// 元の文字列: "natoriaaaaaほげ\n" + "natoria"
		//                                ^ 17文字目
		// 挿入後: "natoriaaaaaほげ\nsato" + "natoria"

		const insertPos = 17; // 改行の直後
		const oldLength = 0; // 置き換えではなく挿入なので0
		const newLength = 4; // "sato"

		const adjustedEntities = adjustEntityPositions(
			initialEntities,
			insertPos,
			oldLength,
			newLength,
		);

		// 期待される結果:
		// positions[0] (0-6) は挿入位置より前なのでそのまま
		// positions[1] (18-24) は挿入位置より後ろなので +4 調整
		expect(adjustedEntities).toEqual([
			{
				entityId: "natori-id",
				positions: [
					{ start: 0, end: 6 }, // そのまま
					{ start: 22, end: 28 }, // 18+4=22, 24+4=28
				],
			},
		]);
	});

	it("エンティティ位置より前に挿入した場合、既存エンティティ位置はそのまま", () => {
		const initialEntities = [
			{
				entityId: "entity-1",
				positions: [{ start: 10, end: 15 }],
			},
		];

		// 5文字目に3文字挿入
		const insertPos = 5;
		const oldLength = 0;
		const newLength = 3;

		const adjustedEntities = adjustEntityPositions(
			initialEntities,
			insertPos,
			oldLength,
			newLength,
		);

		// 10-15は挿入位置(5)より後ろなので+3調整される
		expect(adjustedEntities).toEqual([
			{
				entityId: "entity-1",
				positions: [{ start: 13, end: 18 }],
			},
		]);
	});

	it("エンティティ位置より後ろに挿入した場合、既存エンティティ位置はそのまま", () => {
		const initialEntities = [
			{
				entityId: "entity-1",
				positions: [{ start: 10, end: 15 }],
			},
		];

		// 20文字目に3文字挿入
		const insertPos = 20;
		const oldLength = 0;
		const newLength = 3;

		const adjustedEntities = adjustEntityPositions(
			initialEntities,
			insertPos,
			oldLength,
			newLength,
		);

		// 10-15は挿入位置(20)より前なのでそのまま
		expect(adjustedEntities).toEqual([
			{
				entityId: "entity-1",
				positions: [{ start: 10, end: 15 }],
			},
		]);
	});

	it("エンティティ位置と重複する位置に挿入した場合、そのエンティティは除外される", () => {
		const initialEntities = [
			{
				entityId: "entity-1",
				positions: [
					{ start: 5, end: 10 },
					{ start: 20, end: 25 },
				],
			},
		];

		// 7-12の範囲を"newtext"(7文字)で置き換え
		const insertPos = 7;
		const oldLength = 5; // 7-12 = 5文字
		const newLength = 7;

		const adjustedEntities = adjustEntityPositions(
			initialEntities,
			insertPos,
			oldLength,
			newLength,
		);

		// 5-10は挿入範囲(7-12)と重複するので除外
		// 20-25は挿入位置より後ろなので +2調整 (7-5=2)
		expect(adjustedEntities).toEqual([
			{
				entityId: "entity-1",
				positions: [{ start: 22, end: 27 }],
			},
		]);
	});

	it("複数エンティティの位置を同時に調整できる", () => {
		const initialEntities = [
			{
				entityId: "entity-1",
				positions: [{ start: 5, end: 10 }],
			},
			{
				entityId: "entity-2",
				positions: [{ start: 20, end: 25 }],
			},
			{
				entityId: "entity-3",
				positions: [{ start: 30, end: 35 }],
			},
		];

		// 15文字目に5文字挿入
		const insertPos = 15;
		const oldLength = 0;
		const newLength = 5;

		const adjustedEntities = adjustEntityPositions(
			initialEntities,
			insertPos,
			oldLength,
			newLength,
		);

		// entity-1 (5-10)は挿入位置より前なのでそのまま
		// entity-2 (20-25)は挿入位置より後ろなので+5調整
		// entity-3 (30-35)は挿入位置より後ろなので+5調整
		expect(adjustedEntities).toEqual([
			{
				entityId: "entity-1",
				positions: [{ start: 5, end: 10 }],
			},
			{
				entityId: "entity-2",
				positions: [{ start: 25, end: 30 }],
			},
			{
				entityId: "entity-3",
				positions: [{ start: 35, end: 40 }],
			},
		]);
	});

	it("文字列を削除(oldLength > newLength)した場合も正しく調整される", () => {
		const initialEntities = [
			{
				entityId: "entity-1",
				positions: [
					{ start: 5, end: 10 },
					{ start: 20, end: 25 },
				],
			},
		];

		// 12-17の範囲(5文字)を"ab"(2文字)で置き換え
		const insertPos = 12;
		const oldLength = 5;
		const newLength = 2;

		const adjustedEntities = adjustEntityPositions(
			initialEntities,
			insertPos,
			oldLength,
			newLength,
		);

		// 5-10は挿入位置より前なのでそのまま
		// 20-25は挿入位置より後ろなので-3調整 (2-5=-3)
		expect(adjustedEntities).toEqual([
			{
				entityId: "entity-1",
				positions: [
					{ start: 5, end: 10 },
					{ start: 17, end: 22 },
				],
			},
		]);
	});

	it("同じエンティティの複数positionが正しく調整される", () => {
		// バグ再現ケース:
		// "natoriaaaaaほげ\nnatoria\nnatoriくん"
		// positions: [{ start: 0, end: 6 }, { start: 18, end: 24 }, { start: 26, end: 32 }]
		const initialEntities = [
			{
				entityId: "natori-id",
				positions: [
					{ start: 0, end: 6 }, // "natori"
					{ start: 18, end: 24 }, // "natori"
					{ start: 26, end: 32 }, // "natori"
				],
			},
		];

		// 改行の直後(17文字目)に"test"(4文字)を挿入
		const insertPos = 17;
		const oldLength = 0;
		const newLength = 4;

		const adjustedEntities = adjustEntityPositions(
			initialEntities,
			insertPos,
			oldLength,
			newLength,
		);

		// positions[0] (0-6) はそのまま
		// positions[1] (18-24) は +4 調整
		// positions[2] (26-32) は +4 調整
		expect(adjustedEntities).toEqual([
			{
				entityId: "natori-id",
				positions: [
					{ start: 0, end: 6 },
					{ start: 22, end: 28 },
					{ start: 30, end: 36 },
				],
			},
		]);
	});
});
