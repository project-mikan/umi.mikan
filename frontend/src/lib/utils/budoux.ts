import { loadDefaultJapaneseParser } from "budoux";
import { get } from "svelte/store";
import { budouxEnabled } from "$lib/budoux-store";

// BudouXの日本語パーサーのインスタンス
const parser = loadDefaultJapaneseParser();

/**
 * 日本語テキストにBudouXを適用して、自然な改行位置にタグを挿入する
 * BudouXが無効の場合はテキストをそのまま返す
 * @param text 処理対象のテキスト
 * @returns BudouXで処理されたHTML文字列、または元のテキスト
 */
export function applyBudouX(text: string): string {
	const isEnabled = get(budouxEnabled);
	if (!isEnabled) {
		return text;
	}
	return parser.translateHTMLString(text);
}
