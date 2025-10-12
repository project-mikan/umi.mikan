import type { DiaryEntityOutput } from "$lib/grpc/diary/diary_pb";
import type { Entity } from "$lib/grpc/entity/entity_pb";

/**
 * diaryEntitiesから無効なエンティティ（テキストが名前やエイリアスと一致しないもの）を除外
 */
export function validateDiaryEntities(
	content: string,
	diaryEntities: DiaryEntityOutput[],
	allEntities: Entity[],
): DiaryEntityOutput[] {
	if (
		!diaryEntities ||
		diaryEntities.length === 0 ||
		!allEntities ||
		allEntities.length === 0
	) {
		return diaryEntities;
	}

	const validatedEntities: DiaryEntityOutput[] = [];

	for (const diaryEntity of diaryEntities) {
		// このエンティティIDに対応するEntityを取得
		const entity = allEntities.find((e) => e.id === diaryEntity.entityId);
		if (!entity) {
			// エンティティが見つからない場合はスキップ
			continue;
		}

		// 有効なテキスト（名前とエイリアス）を収集
		const validTexts = [entity.name];
		for (const alias of entity.aliases) {
			validTexts.push(alias.alias);
		}

		// 各positionをチェックし、有効なもののみを残す
		const validPositions = diaryEntity.positions.filter((position) => {
			const text = content.substring(position.start, position.end);
			// textが有効なテキストのいずれかと完全一致するかチェック
			return validTexts.some((validText) => validText === text);
		});

		// 有効なpositionが1つ以上ある場合のみ追加
		if (validPositions.length > 0) {
			validatedEntities.push({
				...diaryEntity,
				positions: validPositions,
			});
		}
	}

	return validatedEntities;
}

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
