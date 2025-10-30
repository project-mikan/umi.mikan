/**
 * HTML/テキスト変換ユーティリティ
 */

/**
 * HTMLをプレーンテキストに変換
 * @param html HTML文字列
 * @returns プレーンテキスト
 */
export function htmlToPlainText(html: string): string {
	// SSR時はシンプルな正規表現処理
	if (typeof document === "undefined") {
		return html.replace(/<br\s*\/?>/gi, "\n").replace(/<[^>]+>/g, "");
	}

	const tempDiv = document.createElement("div");
	tempDiv.innerHTML = html;

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
