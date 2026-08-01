/**
 * HTML/テキスト変換ユーティリティ
 */

import { CURSOR_ANCHOR_ZWS } from "./cursor-utils";

/**
 * HTMLをプレーンテキストに変換
 * @param html HTML文字列
 * @returns プレーンテキスト
 */
export function htmlToPlainText(html: string): string {
  // SSR時はシンプルな正規表現処理
  // カーソルアンカーはブラウザ側のcontenteditable操作でのみ挿入されるため、
  // SSR経路ではそもそも混入し得ない。ユーザーが入力・貼り付けした本物の
  // ゼロ幅スペースまで消してしまわないよう、ここでは除去しない。
  if (typeof document === "undefined") {
    return html.replace(/<br\s*\/?>/gi, "\n").replace(/<[^>]+>/g, "");
  }

  const tempDiv = document.createElement("div");
  tempDiv.innerHTML = html;

  // カーソル位置保持用に挿入されるカーソルアンカー（<br>の直後に挿入される、テキスト
  // ノード先頭のゼロ幅スペース1文字）を除去する（Textarea.svelteのcursor-utils.tsが、
  // <br>直後への入力位置がブラウザによって意図せずずれる問題を避けるためのカーソル
  // アンカーとして挿入している）。<br>の直後にあるテキストノードの先頭1文字だけを
  // ピンポイントで対象にすることで、ユーザーが入力・貼り付けした「テキストの一部としての
  // ゼロ幅スペース」（<br>直後以外にあるもの、または他の文字の後に続くもの）とを区別し、
  // 後者は保持する。
  const brElementsForAnchorCleanup = tempDiv.querySelectorAll("br");
  for (const br of Array.from(brElementsForAnchorCleanup)) {
    const next = br.nextSibling;
    if (
      next?.nodeType === Node.TEXT_NODE &&
      (next.textContent || "").startsWith(CURSOR_ANCHOR_ZWS)
    ) {
      next.textContent = (next.textContent || "").slice(
        CURSOR_ANCHOR_ZWS.length,
      );
    }
  }

  // <br>タグを改行文字に変換
  const brElements = tempDiv.querySelectorAll("br");
  for (const br of Array.from(brElements)) {
    const newline = document.createTextNode("\n");
    br.parentNode?.replaceChild(newline, br);
  }

  // <p>タグと<div>タグの後に改行を追加
  const pElements = tempDiv.querySelectorAll("p");
  for (const p of Array.from(pElements)) {
    const newline = document.createTextNode("\n");
    if (p.nextSibling) {
      p.parentNode?.insertBefore(newline, p.nextSibling);
    }
  }

  const divElements = tempDiv.querySelectorAll("div");
  for (const div of Array.from(divElements)) {
    const newline = document.createTextNode("\n");
    if (div.nextSibling) {
      div.parentNode?.insertBefore(newline, div.nextSibling);
    }
  }

  // <li>タグの処理
  const liElements = tempDiv.querySelectorAll("li");
  for (const li of Array.from(liElements)) {
    const bullet = document.createTextNode("• ");
    li.insertBefore(bullet, li.firstChild);
    const newline = document.createTextNode("\n");
    if (li.nextSibling) {
      li.parentNode?.insertBefore(newline, li.nextSibling);
    }
  }

  let plainText = tempDiv.textContent || tempDiv.innerText || "";

  // 複雑なHTMLの場合のみクリーンアップ
  const hasComplexHTML = /<(?!br\s*\/?>)[^>]+>/.test(html);
  if (hasComplexHTML) {
    plainText = plainText.replace(/^\s+|\s+$/g, "").replace(/[ \t]+/g, " ");
  }

  return plainText;
}

/**
 * HTMLエスケープ
 * @param text プレーンテキスト
 * @returns エスケープされたHTML
 */
export function escapeHtml(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

/**
 * プレーンテキストをHTMLに変換（改行を<br>に）
 * @param text プレーンテキスト
 * @returns HTML
 */
export function plainTextToHtml(text: string): string {
  return escapeHtml(text).replace(/\n/g, "<br>");
}
