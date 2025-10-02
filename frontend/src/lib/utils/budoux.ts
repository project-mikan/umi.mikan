import { loadDefaultJapaneseParser } from "budoux";

// BudouXの日本語パーサーのインスタンス
const parser = loadDefaultJapaneseParser();

/**
 * 日本語テキストにBudouXを適用して、自然な改行位置にタグを挿入する
 * @param text 処理対象のテキスト
 * @returns BudouXで処理されたHTML文字列
 */
export function applyBudouX(text: string): string {
	return parser.translateHTMLString(text);
}
