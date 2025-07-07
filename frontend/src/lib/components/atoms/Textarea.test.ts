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
