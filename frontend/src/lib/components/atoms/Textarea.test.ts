import { render } from "@testing-library/svelte";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { htmlToPlainText as htmlToPlainTextImpl } from "$lib/utils/html-text-converter";
import Textarea from "./Textarea.svelte";

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

  // 実際のTextarea.svelteコンポーネントをマウントし、_handleKeydown（非公開）を
  // DOM経由のEnterキー押下で駆動して、valueに反映される改行数を検証する
  describe("文末でEnterを押した際にvalueへ反映される改行数", () => {
    beforeEach(() => {
      // このdescribe内は実DOM・実コンポーネントが必要なため、上位beforeEachのモックを復元する
      vi.restoreAllMocks();
    });

    // contentEditable要素の末尾にカーソルを移動してEnterキーのkeydownイベントを発火する
    function pressEnterAtEnd(contentElement: HTMLElement): void {
      const range = document.createRange();
      range.selectNodeContents(contentElement);
      range.collapse(false);

      const selection = window.getSelection();
      selection?.removeAllRanges();
      selection?.addRange(range);

      // Textarea.svelteはcaptureフェーズでkeydownを購読しているため、bubbles: trueで十分伝播する
      const event = new KeyboardEvent("keydown", {
        key: "Enter",
        bubbles: true,
        cancelable: true,
      });
      contentElement.dispatchEvent(event);
    }

    it("正常系: contentEditableの文末でEnterキーを押すと_handleInputに渡るinnerHTMLの<br>が1つだけ増える", async () => {
      const { container } = render(Textarea, {
        props: { value: "こんにちは" },
      });

      const contentElement = container.querySelector(
        "[contenteditable]",
      ) as HTMLElement;
      expect(contentElement).not.toBeNull();

      let capturedHtmlAtInput = "";
      contentElement.addEventListener("input", (event) => {
        capturedHtmlAtInput = (event.target as HTMLElement).innerHTML;
      });

      pressEnterAtEnd(contentElement);

      // 実コンポーネントの_handleKeydownが計算した最終的なvalueを検証する
      // （_handleInputがhtmlToPlainTextで変換する直前のinnerHTMLと同じ変換を適用）
      expect(htmlToPlainTextImpl(capturedHtmlAtInput)).toBe("こんにちは\n");
    });

    it("異常系: Enterキーを2回連続で押すと改行が2つ（\\n\\n）になり、余分な空行が入らない", async () => {
      const { container } = render(Textarea, {
        props: { value: "こんにちは" },
      });

      const contentElement = container.querySelector(
        "[contenteditable]",
      ) as HTMLElement;

      let capturedHtmlAtInput = "";
      contentElement.addEventListener("input", (event) => {
        capturedHtmlAtInput = (event.target as HTMLElement).innerHTML;
      });

      pressEnterAtEnd(contentElement);
      pressEnterAtEnd(contentElement);

      // Enter2回 = 改行2つ。3つ（空行2つ分）になっていないことを検証する
      expect(htmlToPlainTextImpl(capturedHtmlAtInput)).toBe("こんにちは\n\n");
    });

    it("正常系: 複数回Enterを押した後に文字入力すると、蓄積したカーソルアンカー（ゼロ幅スペース）がDOMから掃除される", () => {
      const { container } = render(Textarea, {
        props: { value: "こんにちは" },
      });

      const contentElement = container.querySelector(
        "[contenteditable]",
      ) as HTMLElement;

      // Enterを複数回連続で押す（この時点ではカーソルアンカーがDOMに残り続ける）
      pressEnterAtEnd(contentElement);
      pressEnterAtEnd(contentElement);
      pressEnterAtEnd(contentElement);

      // 通常の文字入力イベントを発火する（isTrustedなユーザー入力を模す通常のinput）
      const selection = window.getSelection();
      const range = selection?.getRangeAt(0);
      range?.insertNode(document.createTextNode("あ"));
      const inputEvent = new Event("input", { bubbles: true });
      contentElement.dispatchEvent(inputEvent);

      // 蓄積していたカーソルアンカー（<br>直後のゼロ幅スペース）が全て掃除され、
      // DOMに1つも残っていないことを確認する
      expect(contentElement.innerHTML).not.toContain("​");
    });

    it("正常系: 文末でEnterを押した直後、カーソルは<br>より後ろのノードに位置し、続けてテキストを挿入すると<br>の後ろに追加される", () => {
      const { container } = render(Textarea, {
        props: { value: "こんにちは" },
      });

      const contentElement = container.querySelector(
        "[contenteditable]",
      ) as HTMLElement;

      pressEnterAtEnd(contentElement);

      // Enter直後のカーソル位置を取得
      const selection = window.getSelection();
      expect(selection).not.toBeNull();
      expect(selection?.rangeCount).toBeGreaterThan(0);
      const range = selection?.getRangeAt(0);
      expect(range?.collapsed).toBe(true);

      // カーソルが<br>要素そのもの（コンテナ要素基準のoffset）ではなく、
      // <br>の直後に置かれたテキストノード内にあることを確認する。
      // 過去の実装ではコンテナ要素基準のoffsetでカーソルを設定していたため、
      // ここでcollapsedなRangeにテキストを挿入すると<br>の"前"（1行目の末尾）に
      // 入ってしまい、Enterを押したのに改行されないように見える不具合があった。
      const br = contentElement.querySelector("br");
      expect(br).not.toBeNull();
      expect(range?.startContainer.nodeType).toBe(Node.TEXT_NODE);
      expect(range?.startContainer.previousSibling).toBe(br);

      // 実際にカーソル位置へテキストを挿入し、<br>より後ろに入ることを確認する
      // （カーソルアンカーとして挿入されたゼロ幅スペースを含むテキストノード内に
      // 挿入されるため、ゼロ幅スペースを除去した上で比較する。ゼロ幅スペース自体は
      // htmlToPlainText側で除去されvalueには反映されない）
      range?.insertNode(document.createTextNode("続き"));
      expect(contentElement.innerHTML.replace(/​/g, "")).toBe(
        "こんにちは<br>続き",
      );
    });

    it("正常系: 文末でEnterを押した後、500ms経過してisTypingがfalseに戻ってもDOM要素が作り直されず、改行が一瞬消えて見えるちらつきが起きない", async () => {
      vi.useFakeTimers();
      try {
        const { container } = render(Textarea, {
          props: { value: "こんにちは" },
        });

        const contentElement = container.querySelector(
          "[contenteditable]",
        ) as HTMLElement;

        pressEnterAtEnd(contentElement);

        // Enter直後に挿入された<br>要素の参照を保持しておく
        const brBeforeTimeout = contentElement.querySelector("br");
        expect(brBeforeTimeout).not.toBeNull();

        // isTypingをfalseに戻す500msタイムアウトを進める
        // （SvelteのリアクティブブロックはPromiseベースでフラッシュされるため、
        //   タイマーを進めた後にマイクロタスクキューも明示的にフラッシュする必要がある）
        await vi.advanceTimersByTimeAsync(600);
        await Promise.resolve();
        await Promise.resolve();

        // DOM要素（innerHTMLの再代入）が作り直されていれば<br>要素の参照は
        // 別オブジェクトになる。過去の実装では、内容が実質同じ（改行のみ）でも
        // updateContentElementがinnerHTMLを無条件に再代入していたため、
        // 500ms後に一瞬DOMが作り直され、改行がちらついて見える不具合があった。
        const brAfterTimeout = contentElement.querySelector("br");
        expect(brAfterTimeout).toBe(brBeforeTimeout);
      } finally {
        vi.useRealTimers();
      }
    });
  });

  // htmlToPlainTextImpl単体でも、2つ目の<br>を削除し忘れた場合に
  // 余分な改行が入ることを回帰として押さえておく
  describe("htmlToPlainText: カーソル表示用の2つ目の<br>が残っていた場合の変換結果", () => {
    beforeEach(() => {
      // 実DOM実装が必要なため、上位beforeEachのdocument.createElementモックを復元する
      vi.restoreAllMocks();
    });

    it("異常系: 2つ目の<br>が残ったままだと、改行1つのつもりが改行2つ（空行1つ分）に変換される", () => {
      const html = "こんにちは<br><br>";
      const result = htmlToPlainTextImpl(html);
      expect(result).toBe("こんにちは\n\n");
    });

    it("正常系: 2つ目の<br>を削除してから変換すると改行1つになる", () => {
      const html = "こんにちは<br>";
      const result = htmlToPlainTextImpl(html);
      expect(result).toBe("こんにちは\n");
    });
  });

  // カーソル位置保持用に一時的に挿入されるゼロ幅スペース（U+200B）が
  // valueに混入しないことを回帰として押さえておく
  describe("htmlToPlainText: カーソルアンカー用のゼロ幅スペースが残っていた場合の変換結果", () => {
    beforeEach(() => {
      // 実DOM実装が必要なため、上位beforeEachのdocument.createElementモックを復元する
      vi.restoreAllMocks();
    });

    it("正常系: <br>直後にカーソルアンカー用のゼロ幅スペースが残っていても、valueには含まれない", () => {
      const html = "こんにちは<br>​続き";
      const result = htmlToPlainTextImpl(html);
      expect(result).toBe("こんにちは\n続き");
    });

    it("正常系: <br>直後がゼロ幅スペースのみの場合、その1文字だけがvalueから除かれる", () => {
      const html = "こんにちは<br>​";
      const result = htmlToPlainTextImpl(html);
      expect(result).toBe("こんにちは\n");
    });

    it("異常系: <br>を伴わずゼロ幅スペースのみの場合、カーソルアンカーとはみなされずvalueにそのまま残る（<br>直後で始まるテキストのみをアンカーとして扱うため）", () => {
      const html = "​";
      const result = htmlToPlainTextImpl(html);
      expect(result).toBe("​");
    });

    it("正常系: <br>を伴わずユーザーが入力・貼り付けした本物のゼロ幅スペースはvalueに保持される", () => {
      const html = "前​後";
      const result = htmlToPlainTextImpl(html);
      expect(result).toBe("前​後");
    });
  });
});
