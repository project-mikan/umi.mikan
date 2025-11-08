/**
 * ハイライト情報の型定義
 */
export interface DiaryHighlight {
	start: number;
	end: number;
	text: string;
}

/**
 * ハイライトデータ（APIレスポンス）
 */
export interface HighlightData {
	highlights: DiaryHighlight[];
	createdAt: number;
	updatedAt: number;
}
