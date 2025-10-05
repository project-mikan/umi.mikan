import { create } from "@bufbuild/protobuf";
import {
	DiaryEntityInputSchema,
	type DiaryEntityInput,
} from "$lib/grpc/diary/diary_pb";
import { EntityCategory, PositionSchema } from "$lib/grpc/entity/entity_pb";
import { listEntities } from "$lib/server/entity-api";

/**
 * contentからentity/aliasを抽出してDiaryEntityInputを作成
 */
export async function extractEntitiesFromContent(
	content: string,
	accessToken: string,
): Promise<DiaryEntityInput[]> {
	// 全entityを取得
	const response = await listEntities({
		category: EntityCategory.PEOPLE,
		allCategories: true,
		accessToken,
	});

	const entities = response.entities;

	// 全entity/aliasのリストを作成
	const allCandidates: {
		entityId: string;
		text: string;
	}[] = [];

	for (const entity of entities) {
		allCandidates.push({
			entityId: entity.id,
			text: entity.name,
		});
		for (const alias of entity.aliases) {
			allCandidates.push({
				entityId: entity.id,
				text: alias.alias,
			});
		}
	}

	// contentの各位置で最長の前方一致を探す
	const matches: {
		entityId: string;
		start: number;
		end: number;
		text: string;
	}[] = [];

	let i = 0;
	while (i < content.length) {
		// この位置から始まる全ての候補をチェック
		let longestMatch: {
			entityId: string;
			text: string;
			length: number;
		} | null = null;

		for (const candidate of allCandidates) {
			// この位置から候補の文字列が始まるかチェック
			if (content.substring(i, i + candidate.text.length) === candidate.text) {
				// より長い一致があれば更新
				if (!longestMatch || candidate.text.length > longestMatch.length) {
					longestMatch = {
						entityId: candidate.entityId,
						text: candidate.text,
						length: candidate.text.length,
					};
				}
			}
		}

		if (longestMatch) {
			// 一致が見つかった場合
			matches.push({
				entityId: longestMatch.entityId,
				start: i,
				end: i + longestMatch.length,
				text: longestMatch.text,
			});
			i += longestMatch.length; // 一致した分だけ進む
		} else {
			i++; // 1文字進む
		}
	}

	// entityごとにグループ化してDiaryEntityInputを作成
	const entityMatches = new Map<
		string,
		{ start: number; end: number; text: string }[]
	>();

	for (const match of matches) {
		if (!entityMatches.has(match.entityId)) {
			entityMatches.set(match.entityId, []);
		}
		entityMatches.get(match.entityId)?.push({
			start: match.start,
			end: match.end,
			text: match.text,
		});
	}

	const diaryEntities: DiaryEntityInput[] = [];

	for (const [entityId, positions] of entityMatches) {
		const positionMessages = positions.map((pos) =>
			create(PositionSchema, {
				start: pos.start,
				end: pos.end,
			}),
		);

		diaryEntities.push(
			create(DiaryEntityInputSchema, {
				entityId: entityId,
				positions: positionMessages,
				usedText: positions[0].text, // 最初に見つかったテキストを使用
			}),
		);
	}

	return diaryEntities;
}
