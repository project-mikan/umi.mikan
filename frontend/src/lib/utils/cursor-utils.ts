/**
 * contenteditable要素のカーソル位置管理ユーティリティ
 */

/**
 * <br>直後のカーソル位置を安定させるために挿入するカーソルアンカー用のゼロ幅スペース（U+200B）。
 * DOM上にのみ一時的に存在し、value（プレーンテキスト）には反映されない。
 * リテラルが複数箇所に散らばると改変時に不一致を起こしやすいため、ここで一元管理する。
 */
export const CURSOR_ANCHOR_ZWS = "​";

/**
 * contenteditable要素内のテキストオフセット位置を取得
 * @param root ルート要素
 * @param node 対象ノード
 * @param offset ノード内のオフセット
 * @returns テキストオフセット位置
 */
export function getTextOffset(root: Node, node: Node, offset: number): number {
  let textOffset = 0;

  // カーソルアンカー（<br>直後に挿入されるゼロ幅スペース1文字）はDOM上にのみ一時的に
  // 存在し、valueには反映されない。「直前の兄弟ノードが<br>」かつ「テキストノードの
  // 先頭がゼロ幅スペース」の場合のみアンカーとみなしてその1文字分をオフセット計算から
  // 除外する（<br>直後以外にあるゼロ幅スペースはユーザー入力由来の可能性があるため対象外）。
  // これを数えてしまうと、アンカーがまだ除去されていないタイミングでcursorPosを取得した
  // 場合に、アンカー除去後のDOM（またはZWSを含まないvalue由来のHTML）に対して
  // restoreCursorPositionした際、実際のカーソル位置より1つ後ろにずれてしまう。
  function hasLeadingAnchor(textNode: Node): boolean {
    return (
      textNode.previousSibling?.nodeName === "BR" &&
      (textNode.textContent || "").startsWith(CURSOR_ANCHOR_ZWS)
    );
  }

  function textNodeLength(textNode: Node): number {
    const content = textNode.textContent || "";
    return hasLeadingAnchor(textNode) ? content.length - 1 : content.length;
  }

  function traverse(currentNode: Node): number | null {
    if (currentNode === node) {
      if (currentNode.nodeType === Node.TEXT_NODE) {
        const anchorAdjustment =
          hasLeadingAnchor(currentNode) && offset > 0 ? 1 : 0;
        return textOffset + offset - anchorAdjustment;
      }
      if (currentNode.nodeType === Node.ELEMENT_NODE) {
        const children = Array.from(currentNode.childNodes);
        for (let i = 0; i < Math.min(offset, children.length); i++) {
          const child = children[i];
          if (child.nodeType === Node.TEXT_NODE) {
            textOffset += textNodeLength(child);
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
      textOffset += textNodeLength(currentNode);
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
      return textNodeLength(node);
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
          // BRの直後にカーソルを置く際、コンテナ要素基準のoffset（parent, index）で
          // Rangeを設定すると、ブラウザによっては直後のキー入力がBRの"前"ではなく
          // "後ろ"に挿入されてしまうことがある（Chromeで確認済み）。これを避けるため、
          // BRの直後にゼロ幅スペース（U+200B）を1文字持つテキストノードを挿入し、
          // そのテキストノード内（offset 1、＝ゼロ幅スペースの直後）をカーソル位置として使う。
          // 空文字列のテキストノードだとブラウザが正規化時に削除してしまいカーソル位置が
          // 失われることがあったため、削除されない1文字のゼロ幅スペースを使う。
          // ゼロ幅スペースはhtmlToPlainText側で除去されるためvalueには反映されない。
          const anchor = document.createTextNode(CURSOR_ANCHOR_ZWS);
          node.parentNode?.insertBefore(anchor, node.nextSibling);
          targetNode = anchor;
          targetOffset = 1;
          return true;
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
    } catch (error: unknown) {
      if (error instanceof Error) {
        console.error("Failed to restore cursor position:", error.message);
      } else {
        console.error("Failed to restore cursor position:", error);
      }
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
    } catch (error: unknown) {
      if (error instanceof Error) {
        console.error("Failed to create range at text offset:", error.message);
      } else {
        console.error("Failed to create range at text offset:", error);
      }
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
  } catch (error: unknown) {
    // カーソル復元の最終手段が失敗した場合は何もしない
    if (error instanceof Error) {
      console.error("Failed to fallback to end:", error.message);
    }
  }
}

/**
 * contentElement内に残存するカーソルアンカー用ゼロ幅スペース（テキストノード先頭の1文字）を
 * 全て除去する。<br>挿入直後のカーソル安定化のためだけに使われる一時的な文字であり、役目を
 * 終えたらDOMに残し続けてはいけない（際限なく蓄積し、htmlToPlainTextでのユーザー由来ZWSとの
 * 区別も不可能になるため）。現在のカーソル位置を保ったまま除去する。
 * @param contentElement contenteditable要素
 */
export function cleanupCursorAnchors(contentElement: HTMLDivElement): void {
  if (typeof window === "undefined") return;

  const walker = document.createTreeWalker(
    contentElement,
    NodeFilter.SHOW_TEXT,
  );
  const anchorNodes: Text[] = [];
  let current = walker.nextNode();
  while (current) {
    if (
      current.previousSibling?.nodeName === "BR" &&
      (current.textContent || "").startsWith(CURSOR_ANCHOR_ZWS)
    ) {
      anchorNodes.push(current as Text);
    }
    current = walker.nextNode();
  }

  if (anchorNodes.length === 0) return;

  const selection = window.getSelection();
  let cursorPos: number | null = null;
  if (selection && selection.rangeCount > 0) {
    const range = selection.getRangeAt(0);
    if (contentElement.contains(range.startContainer)) {
      cursorPos = getTextOffset(
        contentElement,
        range.startContainer,
        range.startOffset,
      );
    }
  }

  for (const textNode of anchorNodes) {
    textNode.textContent = (textNode.textContent || "").slice(
      CURSOR_ANCHOR_ZWS.length,
    );
  }

  // ここでrestoreCursorPositionは使わない。targetPosがちょうど<br>の位置に一致する場合、
  // restoreCursorPositionは（意図的に）新しいカーソルアンカーを挿入し直してしまうため、
  // 「アンカーを除去する」はずのこの関数がアンカーを再生成する無限ループ状態になる。
  // createRangeAtTextOffsetはアンカーを作らずRangeを返すだけなので、除去後の位置復元には
  // こちらを使う。
  if (cursorPos !== null) {
    const range = createRangeAtTextOffset(contentElement, cursorPos);
    if (range && selection) {
      selection.removeAllRanges();
      selection.addRange(range);
    }
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
