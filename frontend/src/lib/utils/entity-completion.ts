/**
 * エンティティ補完ユーティリティ
 */

import type { Entity } from "$lib/grpc/entity/entity_pb";

export interface FlatSuggestion {
	entity: Entity;
	text: string;
	isAlias: boolean;
}

export interface EntityPosition {
	start: number;
	end: number;
}

export interface SelectedEntity {
	entityId: string;
	positions: EntityPosition[];
}

/**
 * エンティティデータを取得してフラット化
 * @returns エンティティの配列とフラット化された候補リスト
 */
export async function loadAllEntities(): Promise<{
	entities: Entity[];
	flatEntities: FlatSuggestion[];
}> {
	try {
		const response = await fetch("/api/entities/search?q=");
		const data = await response.json();
		const entities: Entity[] = data.entities || [];

		// フラット化
		const flatEntities: FlatSuggestion[] = [];
		for (const entity of entities) {
			flatEntities.push({ entity, text: entity.name, isAlias: false });
			for (const alias of entity.aliases) {
				flatEntities.push({ entity, text: alias.alias, isAlias: true });
			}
		}

		return { entities, flatEntities };
	} catch (err) {
		console.error("Failed to load all entities:", err);
		return { entities: [], flatEntities: [] };
	}
}

/**
 * 前方一致でエンティティをフィルタリング
 * @param query 検索クエリ
 * @param allFlatEntities 全てのフラット化エンティティ
 * @returns フィルタリングされた候補リスト
 */
export function filterEntitiesByPrefix(
	query: string,
	allFlatEntities: FlatSuggestion[],
): FlatSuggestion[] {
	const lowerQuery = query.toLowerCase();
	const matchingEntityIds = new Set<string>();

	for (const flat of allFlatEntities) {
		if (flat.text.toLowerCase().startsWith(lowerQuery)) {
			matchingEntityIds.add(flat.entity.id);
		}
	}

	// マッチしたエンティティの全バリエーション（名前+エイリアス）を含める
	return allFlatEntities.filter((flat) =>
		matchingEntityIds.has(flat.entity.id),
	);
}

/**
 * 最も前方一致する候補のインデックスを取得
 * @param query 検索クエリ
 * @param suggestions 候補リスト
 * @returns インデックス
 */
export function getBestMatchIndex(
	query: string,
	suggestions: FlatSuggestion[],
): number {
	if (!query || suggestions.length === 0) return 0;

	const lowerQuery = query.toLowerCase();

	// 先頭一致するものを探す
	for (let i = 0; i < suggestions.length; i++) {
		if (suggestions[i].text.toLowerCase().startsWith(lowerQuery)) {
			return i;
		}
	}

	return 0;
}

/**
 * カーソル位置から最長一致するエンティティを見つける
 * @param searchText カーソル前のテキスト
 * @param allFlatEntities 全てのフラット化エンティティ
 * @param cursorPos カーソル位置
 * @returns 一致情報またはnull
 */
export function findLongestMatch(
	searchText: string,
	allFlatEntities: FlatSuggestion[],
	cursorPos: number,
): { word: string; startPos: number } | null {
	// 後方から2文字以上の部分文字列を試す
	for (let len = searchText.length; len >= 2; len--) {
		const substring = searchText.substring(searchText.length - len);

		// このsubstringで始まるエンティティがあるかチェック
		const hasMatch = allFlatEntities.some((flat) =>
			flat.text.toLowerCase().startsWith(substring.toLowerCase()),
		);

		if (hasMatch) {
			return {
				word: substring,
				startPos: cursorPos - len,
			};
		}
	}

	return null;
}

/**
 * positionリストを調整（エンティティ選択後のテキスト置換に対応）
 * @param positions 元のpositionリスト
 * @param replaceStart 置換開始位置
 * @param replaceEnd 置換終了位置
 * @param lengthDiff 長さの差分
 * @returns 調整されたpositionリスト
 */
export function adjustPositions(
	positions: EntityPosition[],
	replaceStart: number,
	replaceEnd: number,
	lengthDiff: number,
): EntityPosition[] {
	return positions
		.map((pos) => {
			// 置き換え範囲の前のpositionはそのまま
			if (pos.end <= replaceStart) {
				return pos;
			}
			// 置き換え範囲と完全に重複するpositionは除外
			if (pos.start >= replaceStart && pos.end <= replaceEnd) {
				return null;
			}
			// 置き換え範囲とpositionが部分的に重複する場合も除外
			if (
				(pos.start < replaceStart &&
					pos.end > replaceStart &&
					pos.end <= replaceEnd) ||
				(pos.start >= replaceStart &&
					pos.start < replaceEnd &&
					pos.end > replaceEnd)
			) {
				return null;
			}
			// 置き換え範囲より後ろのpositionは調整
			if (pos.start >= replaceEnd) {
				return {
					start: pos.start + lengthDiff,
					end: pos.end + lengthDiff,
				};
			}
			// その他（開始が挿入位置より前で、終了が置き換え範囲より後ろ）
			return {
				start: pos.start,
				end: pos.end + lengthDiff,
			};
		})
		.filter((pos): pos is EntityPosition => pos !== null);
}

/**
 * selectedEntitiesからエンティティハイライトを適用したHTMLを生成
 * @param content プレーンテキスト
 * @param selectedEntities 選択されたエンティティリスト
 * @returns エンティティハイライト付きHTML
 */
export function generateEntityHighlightHTML(
	content: string,
	selectedEntities: SelectedEntity[],
): string {
	if (!selectedEntities || selectedEntities.length === 0) {
		return content.replace(/\n/g, "<br>");
	}

	interface HighlightSegment {
		start: number;
		end: number;
		entityId: string;
	}

	const segments: HighlightSegment[] = [];

	for (const selectedEnt of selectedEntities) {
		for (const position of selectedEnt.positions) {
			segments.push({
				start: position.start,
				end: position.end,
				entityId: selectedEnt.entityId,
			});
		}
	}

	// 開始位置でソート
	segments.sort((a, b) => a.start - b.start);

	// 重複・重なり合ったsegmentをフィルタリング
	const validSegments: HighlightSegment[] = [];
	let lastEnd = 0;

	for (const segment of segments) {
		if (
			segment.start >= lastEnd &&
			segment.start < content.length &&
			segment.end <= content.length &&
			segment.start < segment.end
		) {
			validSegments.push(segment);
			lastEnd = segment.end;
		}
	}

	// HTMLを構築
	let result = "";
	let lastIndex = 0;

	for (const segment of validSegments) {
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
