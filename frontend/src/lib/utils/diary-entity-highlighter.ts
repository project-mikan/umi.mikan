import type { DiaryHighlight } from "$lib/types/highlight";

function escapeHtml(text: string): string {
	return text
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#039;");
}

/**
 * 日記のcontentにdiaryHighlightsと検索キーワードをハイライトしたHTMLを生成
 */
export function highlightEntitiesAndHighlights(
	content: string,
	_diaryEntities: unknown[],
	diaryHighlights: DiaryHighlight[],
	searchKeyword?: string,
): string {
	interface HighlightSegment {
		start: number;
		end: number;
		type: "diary" | "search";
	}

	const segments: HighlightSegment[] = (diaryHighlights ?? []).map(
		(highlight) => ({
			start: highlight.start,
			end: highlight.end,
			type: "diary" as const,
		}),
	);

	// 検索キーワードの出現位置を追加
	if (searchKeyword) {
		const lowerContent = content.toLowerCase();
		const lowerKeyword = searchKeyword.toLowerCase();
		let idx = lowerContent.indexOf(lowerKeyword, 0);
		while (idx !== -1) {
			segments.push({
				start: idx,
				end: idx + searchKeyword.length,
				type: "search",
			});
			idx = lowerContent.indexOf(lowerKeyword, idx + 1);
		}
	}

	if (segments.length === 0) {
		return escapeHtml(content).replace(/\n/g, "<br>");
	}

	// 開始位置でソート（同じ開始位置の場合はsearchを優先）
	segments.sort((a, b) => a.start - b.start || (a.type === "search" ? -1 : 1));

	// HTMLを構築（重複は後続セグメントをスキップ）
	let result = "";
	let lastIndex = 0;

	for (const segment of segments) {
		// 前のセグメントと重複する場合はスキップ
		if (segment.start < lastIndex) continue;

		// segment前のテキスト
		if (lastIndex < segment.start) {
			result += escapeHtml(content.substring(lastIndex, segment.start)).replace(
				/\n/g,
				"<br>",
			);
		}

		const segmentText = escapeHtml(
			content.substring(segment.start, segment.end),
		);
		if (segment.type === "search") {
			// 検索キーワード（オレンジ背景）
			result += `<mark class="bg-orange-300 dark:bg-orange-500 px-1 rounded font-medium">${segmentText}</mark>`;
		} else {
			// AIハイライト（黄色背景）
			result += `<mark class="bg-yellow-300 dark:bg-yellow-600 px-1 rounded font-medium">${segmentText}</mark>`;
		}

		lastIndex = segment.end;
	}

	// 残りのテキスト
	if (lastIndex < content.length) {
		result += escapeHtml(content.substring(lastIndex)).replace(/\n/g, "<br>");
	}

	return result;
}
