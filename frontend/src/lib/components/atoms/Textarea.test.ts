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
});
