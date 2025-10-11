import type { DiaryEntityOutput } from "$lib/grpc/diary/diary_pb";

/**
 * 日記のcontentとdiaryEntitiesから、entity/aliasをハイライトしたHTMLを生成
 */
export function highlightEntities(
	content: string,
	diaryEntities: DiaryEntityOutput[],
): string {
	if (!diaryEntities || diaryEntities.length === 0) {
		// エスケープしてから改行を<br>に変換
		return escapeHtml(content).replace(/\n/g, "<br>");
	}

	// 全てのpositionを収集してソート
	interface HighlightSegment {
		start: number;
		end: number;
		entityId: string;
	}

	const segments: HighlightSegment[] = [];

	for (const diaryEntity of diaryEntities) {
		for (const position of diaryEntity.positions) {
			segments.push({
				start: position.start,
				end: position.end,
				entityId: diaryEntity.entityId,
			});
		}
	}

	// 開始位置でソート
	segments.sort((a, b) => a.start - b.start);

	// HTMLを構築
	let result = "";
	let lastIndex = 0;

	for (const segment of segments) {
		// segment前のテキスト
		if (lastIndex < segment.start) {
			const text = content.substring(lastIndex, segment.start);
			result += escapeHtml(text).replace(/\n/g, "<br>");
		}

		// segmentのテキスト(リンク付き青色)
		const entityText = content.substring(segment.start, segment.end);
		result += `<a href="/entity/${segment.entityId}" class="text-blue-600 dark:text-blue-400 hover:underline">${escapeHtml(entityText)}</a>`;

		lastIndex = segment.end;
	}

	// 残りのテキスト
	if (lastIndex < content.length) {
		const text = content.substring(lastIndex);
		result += escapeHtml(text).replace(/\n/g, "<br>");
	}

	return result;
}

function escapeHtml(text: string): string {
	return text
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#039;");
}
