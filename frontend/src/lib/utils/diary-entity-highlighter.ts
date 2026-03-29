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
 * 日記のcontentにdiaryHighlightsをハイライトしたHTMLを生成
 */
export function highlightEntitiesAndHighlights(
	content: string,
	_diaryEntities: unknown[],
	diaryHighlights: DiaryHighlight[],
): string {
	// ハイライトセグメントを追加
	if (!diaryHighlights || diaryHighlights.length === 0) {
		return escapeHtml(content).replace(/\n/g, "<br>");
	}

	interface HighlightSegment {
		start: number;
		end: number;
	}

	const segments: HighlightSegment[] = diaryHighlights.map((highlight) => ({
		start: highlight.start,
		end: highlight.end,
	}));

	// 開始位置でソート
	segments.sort((a, b) => a.start - b.start);

	// 重複するセグメントをマージ
	const mergedSegments: HighlightSegment[] = [];
	for (const segment of segments) {
		const overlapping = mergedSegments.find(
			(s) =>
				(segment.start >= s.start && segment.start < s.end) ||
				(segment.end > s.start && segment.end <= s.end) ||
				(segment.start <= s.start && segment.end >= s.end),
		);

		if (!overlapping) {
			mergedSegments.push(segment);
		}
	}

	// HTMLを構築
	let result = "";
	let lastIndex = 0;

	for (const segment of mergedSegments) {
		// segment前のテキスト
		if (lastIndex < segment.start) {
			const text = content.substring(lastIndex, segment.start);
			result += escapeHtml(text).replace(/\n/g, "<br>");
		}

		// ハイライト（黄色背景）
		const segmentText = content.substring(segment.start, segment.end);
		result += `<mark class="bg-yellow-300 dark:bg-yellow-600 px-1 rounded font-medium">${escapeHtml(segmentText)}</mark>`;

		lastIndex = segment.end;
	}

	// 残りのテキスト
	if (lastIndex < content.length) {
		const text = content.substring(lastIndex);
		result += escapeHtml(text).replace(/\n/g, "<br>");
	}

	return result;
}
