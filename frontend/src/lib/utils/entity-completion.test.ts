/**
 * entity-completionのテスト
 */
import { describe, expect, it, vi, beforeEach } from "vitest";
import {
	loadAllEntities,
	filterEntitiesByPrefix,
	getBestMatchIndex,
	findLongestMatch,
	adjustPositions,
	generateEntityHighlightHTML,
	type FlatSuggestion,
	type EntityPosition,
	type SelectedEntity,
} from "./entity-completion";
import type { Entity } from "$lib/grpc/entity/entity_pb";

// モックエンティティデータを作成
const createMockEntity = (
	id: string,
	name: string,
	aliases: string[] = [],
): Entity => {
	return {
		id,
		name,
		aliases: aliases.map((alias) => ({ alias })),
	} as Entity;
};

describe("entity-completion", () => {
	describe("loadAllEntities", () => {
		beforeEach(() => {
			// fetchのモックをリセット
			vi.restoreAllMocks();
		});

		it("エンティティデータを取得してフラット化する", async () => {
			const mockData = {
				entities: [
					createMockEntity("1", "テストエンティティ", [
						"エイリアス1",
						"エイリアス2",
					]),
					createMockEntity("2", "サンプル"),
				],
			};

			global.fetch = vi.fn().mockResolvedValue({
				json: async () => mockData,
			});

			const result = await loadAllEntities();

			expect(result.entities).toHaveLength(2);
			expect(result.flatEntities).toHaveLength(4); // 2エンティティ + 2エイリアス
			expect(result.flatEntities[0].text).toBe("テストエンティティ");
			expect(result.flatEntities[0].isAlias).toBe(false);
			expect(result.flatEntities[1].text).toBe("エイリアス1");
			expect(result.flatEntities[1].isAlias).toBe(true);
		});

		it("APIエラー時に空の配列を返す", async () => {
			global.fetch = vi.fn().mockRejectedValue(new Error("API Error"));

			const result = await loadAllEntities();

			expect(result.entities).toEqual([]);
			expect(result.flatEntities).toEqual([]);
		});
	});

	describe("filterEntitiesByPrefix", () => {
		const flatEntities: FlatSuggestion[] = [
			{
				entity: createMockEntity("1", "テスト"),
				text: "テスト",
				isAlias: false,
			},
			{
				entity: createMockEntity("1", "テスト"),
				text: "テストエイリアス",
				isAlias: true,
			},
			{
				entity: createMockEntity("2", "サンプル"),
				text: "サンプル",
				isAlias: false,
			},
			{
				entity: createMockEntity("3", "テストサンプル"),
				text: "テストサンプル",
				isAlias: false,
			},
		];

		it("前方一致でフィルタリングし、同じエンティティの全バリエーションを含める", () => {
			const result = filterEntitiesByPrefix("テス", flatEntities);

			// "テスト"エンティティの名前とエイリアス、"テストサンプル"がマッチ
			expect(result).toHaveLength(3);
			expect(result.some((f) => f.text === "テスト")).toBe(true);
			expect(result.some((f) => f.text === "テストエイリアス")).toBe(true);
			expect(result.some((f) => f.text === "テストサンプル")).toBe(true);
		});

		it("大文字小文字を区別しない（英語）", () => {
			const flatEntitiesEn: FlatSuggestion[] = [
				{
					entity: createMockEntity("1", "Test"),
					text: "Test",
					isAlias: false,
				},
			];
			const result = filterEntitiesByPrefix("test", flatEntitiesEn);
			expect(result.length).toBeGreaterThan(0);
		});

		it("マッチしない場合は空の配列を返す", () => {
			const result = filterEntitiesByPrefix("存在しない", flatEntities);
			expect(result).toEqual([]);
		});

		it("空のクエリですべてを返す", () => {
			const result = filterEntitiesByPrefix("", flatEntities);
			expect(result).toEqual(flatEntities);
		});
	});

	describe("getBestMatchIndex", () => {
		const suggestions: FlatSuggestion[] = [
			{
				entity: createMockEntity("1", "サンプル"),
				text: "サンプル",
				isAlias: false,
			},
			{
				entity: createMockEntity("2", "テスト"),
				text: "テスト",
				isAlias: false,
			},
			{
				entity: createMockEntity("3", "テストデータ"),
				text: "テストデータ",
				isAlias: false,
			},
		];

		it("前方一致する最初の候補のインデックスを返す", () => {
			const index = getBestMatchIndex("テス", suggestions);
			expect(index).toBe(1); // "テスト"
		});

		it("大文字小文字を区別しない（英語）", () => {
			const suggestionsEn: FlatSuggestion[] = [
				{
					entity: createMockEntity("1", "Sample"),
					text: "Sample",
					isAlias: false,
				},
				{
					entity: createMockEntity("2", "Test"),
					text: "Test",
					isAlias: false,
				},
			];
			const index = getBestMatchIndex("test", suggestionsEn);
			expect(index).toBe(1);
		});

		it("マッチしない場合は0を返す", () => {
			const index = getBestMatchIndex("存在しない", suggestions);
			expect(index).toBe(0);
		});

		it("空のクエリで0を返す", () => {
			const index = getBestMatchIndex("", suggestions);
			expect(index).toBe(0);
		});

		it("空の候補リストで0を返す", () => {
			const index = getBestMatchIndex("テスト", []);
			expect(index).toBe(0);
		});
	});

	describe("findLongestMatch", () => {
		const flatEntities: FlatSuggestion[] = [
			{
				entity: createMockEntity("1", "テスト"),
				text: "テスト",
				isAlias: false,
			},
			{
				entity: createMockEntity("2", "テストデータ"),
				text: "テストデータ",
				isAlias: false,
			},
		];

		it("最長一致する部分文字列を見つける", () => {
			const result = findLongestMatch("これはテストデ", flatEntities, 8);

			expect(result).not.toBeNull();
			if (result) {
				expect(result.word).toBe("テストデ");
				expect(result.startPos).toBe(4); // 8 - 4（"テストデ"は4文字）
			}
		});

		it("短い一致を見つける", () => {
			const result = findLongestMatch("これはテス", flatEntities, 6);

			expect(result).not.toBeNull();
			if (result) {
				expect(result.word).toBe("テス");
				expect(result.startPos).toBe(4); // 6 - 2
			}
		});

		it("2文字未満ではマッチしない", () => {
			const result = findLongestMatch("テ", flatEntities, 1);
			expect(result).toBeNull();
		});

		it("マッチしない場合はnullを返す", () => {
			const result = findLongestMatch("存在しない", flatEntities, 5);
			expect(result).toBeNull();
		});
	});

	describe("adjustPositions", () => {
		describe("基本的な調整", () => {
			it("置換範囲の前のpositionはそのまま", () => {
				const positions: EntityPosition[] = [
					{ start: 0, end: 5 },
					{ start: 10, end: 15 },
				];

				const result = adjustPositions(positions, 20, 25, 5);

				expect(result).toEqual([
					{ start: 0, end: 5 },
					{ start: 10, end: 15 },
				]);
			});

			it("置換範囲と完全に重複するpositionは除外", () => {
				const positions: EntityPosition[] = [
					{ start: 0, end: 5 },
					{ start: 10, end: 15 }, // この範囲が置換される
					{ start: 20, end: 25 },
				];

				const result = adjustPositions(positions, 10, 15, 5);

				expect(result).toHaveLength(2);
				expect(result).toEqual([
					{ start: 0, end: 5 },
					{ start: 25, end: 30 }, // 20+5, 25+5
				]);
			});

			it("置換範囲より後ろのpositionは調整される", () => {
				const positions: EntityPosition[] = [
					{ start: 20, end: 25 },
					{ start: 30, end: 35 },
				];

				const result = adjustPositions(positions, 10, 15, 5);

				expect(result).toEqual([
					{ start: 25, end: 30 }, // 20+5, 25+5
					{ start: 35, end: 40 }, // 30+5, 35+5
				]);
			});
		});

		describe("部分重複の処理", () => {
			it("開始が置換範囲の前で、終了が置換範囲内の場合は除外", () => {
				const positions: EntityPosition[] = [
					{ start: 5, end: 12 }, // 10-15の範囲と部分的に重複
				];

				const result = adjustPositions(positions, 10, 15, 3);

				expect(result).toEqual([]);
			});

			it("開始が置換範囲内で、終了が置換範囲の後の場合は除外", () => {
				const positions: EntityPosition[] = [
					{ start: 12, end: 20 }, // 10-15の範囲と部分的に重複
				];

				const result = adjustPositions(positions, 10, 15, 3);

				expect(result).toEqual([]);
			});
		});

		describe("境界条件", () => {
			it("置換範囲の直前で終わるpositionはそのまま", () => {
				const positions: EntityPosition[] = [{ start: 5, end: 10 }];

				const result = adjustPositions(positions, 10, 15, 5);

				expect(result).toEqual([{ start: 5, end: 10 }]);
			});

			it("置換範囲の直後で始まるpositionは調整される", () => {
				const positions: EntityPosition[] = [{ start: 15, end: 20 }];

				const result = adjustPositions(positions, 10, 15, 5);

				expect(result).toEqual([{ start: 20, end: 25 }]);
			});

			it("負のlengthDiff（削除）を正しく処理する", () => {
				const positions: EntityPosition[] = [
					{ start: 0, end: 5 },
					{ start: 20, end: 25 },
				];

				const result = adjustPositions(positions, 10, 15, -5);

				expect(result).toEqual([
					{ start: 0, end: 5 },
					{ start: 15, end: 20 }, // 20-5, 25-5
				]);
			});

			it("複数のpositionが混在する複雑なケース", () => {
				const positions: EntityPosition[] = [
					{ start: 0, end: 5 }, // 前 -> そのまま
					{ start: 8, end: 12 }, // 部分重複 -> 除外
					{ start: 10, end: 15 }, // 完全重複 -> 除外
					{ start: 13, end: 18 }, // 部分重複 -> 除外
					{ start: 20, end: 25 }, // 後 -> 調整
					{ start: 30, end: 35 }, // 後 -> 調整
				];

				const result = adjustPositions(positions, 10, 15, 3);

				expect(result).toEqual([
					{ start: 0, end: 5 },
					{ start: 23, end: 28 }, // 20+3, 25+3
					{ start: 33, end: 38 }, // 30+3, 35+3
				]);
			});
		});

		describe("エッジケース", () => {
			it("空のpositionリストは空を返す", () => {
				const result = adjustPositions([], 10, 15, 5);
				expect(result).toEqual([]);
			});

			it("lengthDiffが0の場合", () => {
				const positions: EntityPosition[] = [
					{ start: 0, end: 5 },
					{ start: 20, end: 25 },
				];

				const result = adjustPositions(positions, 10, 15, 0);

				expect(result).toEqual([
					{ start: 0, end: 5 },
					{ start: 20, end: 25 },
				]);
			});

			it("置換範囲が0の場合（挿入のみ）", () => {
				const positions: EntityPosition[] = [
					{ start: 0, end: 5 },
					{ start: 10, end: 15 },
				];

				const result = adjustPositions(positions, 10, 10, 5);

				expect(result).toEqual([
					{ start: 0, end: 5 },
					{ start: 15, end: 20 }, // 10+5, 15+5
				]);
			});
		});

		describe("連続削除のシミュレーション（累積オフセット）", () => {
			it("連続削除のシナリオを正しく処理する", () => {
				// 初期状態: [0-5], [10-15], [20-25], [30-35]
				// 削除1: [10-15] を削除 (lengthDiff: -5)
				// 削除2: [20-25] を削除（すでに [15-20] に調整済み） (lengthDiff: -5)

				let positions: EntityPosition[] = [
					{ start: 0, end: 5 },
					{ start: 10, end: 15 },
					{ start: 20, end: 25 },
					{ start: 30, end: 35 },
				];

				// 削除1: [10-15]
				positions = adjustPositions(positions, 10, 15, -5);
				expect(positions).toEqual([
					{ start: 0, end: 5 },
					{ start: 15, end: 20 }, // 20-5
					{ start: 25, end: 30 }, // 30-5
				]);

				// 削除2: [15-20] (元の [20-25])
				positions = adjustPositions(positions, 15, 20, -5);
				expect(positions).toEqual([
					{ start: 0, end: 5 },
					{ start: 20, end: 25 }, // 25-5
				]);
			});
		});
	});

	describe("generateEntityHighlightHTML", () => {
		it("エンティティなしの場合は改行のみ変換", () => {
			const result = generateEntityHighlightHTML("テスト\nテキスト", []);
			expect(result).toBe("テスト<br>テキスト");
		});

		it("単一エンティティをハイライトする", () => {
			const selectedEntities: SelectedEntity[] = [
				{
					entityId: "entity-1",
					positions: [{ start: 0, end: 3 }], // "テスト"は3文字
				},
			];

			const result = generateEntityHighlightHTML(
				"テストテキスト",
				selectedEntities,
			);

			expect(result).toContain('<a href="/entity/entity-1"');
			expect(result).toContain("テスト");
			expect(result).toContain("テキスト");
		});

		it("複数のエンティティをハイライトする", () => {
			const selectedEntities: SelectedEntity[] = [
				{
					entityId: "entity-1",
					positions: [{ start: 0, end: 3 }], // "テスト"は3文字
				},
				{
					entityId: "entity-2",
					positions: [{ start: 3, end: 7 }], // "テキスト"は4文字
				},
			];

			const result = generateEntityHighlightHTML(
				"テストテキスト",
				selectedEntities,
			);

			expect(result).toContain('<a href="/entity/entity-1"');
			expect(result).toContain('<a href="/entity/entity-2"');
		});

		it("同じエンティティの複数出現をハイライトする", () => {
			const selectedEntities: SelectedEntity[] = [
				{
					entityId: "entity-1",
					positions: [
						{ start: 0, end: 3 }, // 最初の"テスト"は3文字
						{ start: 6, end: 9 }, // 2番目の"テスト"（0-2: テスト, 3-5: abc, 6-8: テスト）
					],
				},
			];

			const result = generateEntityHighlightHTML(
				"テストabcテスト",
				selectedEntities,
			);

			const matches = result.match(/<a href="\/entity\/entity-1"/g);
			expect(matches).toHaveLength(2);
		});

		it("HTMLエスケープを正しく行う", () => {
			const selectedEntities: SelectedEntity[] = [
				{
					entityId: "entity-1",
					positions: [{ start: 0, end: 8 }], // "<script>"は8文字
				},
			];

			const result = generateEntityHighlightHTML(
				"<script>alert('xss')</script>",
				selectedEntities,
			);

			expect(result).toContain("&lt;script");
			expect(result).toContain("&gt;");
			expect(result).not.toContain("<script>");
		});

		it("改行を含むテキストを正しく処理する", () => {
			const selectedEntities: SelectedEntity[] = [
				{
					entityId: "entity-1",
					positions: [{ start: 0, end: 3 }], // "テスト"は3文字
				},
			];

			const result = generateEntityHighlightHTML(
				"テスト\nテキスト",
				selectedEntities,
			);

			expect(result).toContain('<a href="/entity/entity-1"');
			expect(result).toContain("<br>");
		});

		it("重複するpositionをフィルタリングする", () => {
			const selectedEntities: SelectedEntity[] = [
				{
					entityId: "entity-1",
					positions: [
						{ start: 0, end: 3 }, // "テスト"
						{ start: 2, end: 5 }, // "ストテ" - 重複
					],
				},
			];

			const result = generateEntityHighlightHTML(
				"テストテキスト",
				selectedEntities,
			);

			// 最初のpositionのみが適用される
			const matches = result.match(/<a href="\/entity\/entity-1"/g);
			expect(matches).toHaveLength(1);
		});

		it("範囲外のpositionを無視する", () => {
			const selectedEntities: SelectedEntity[] = [
				{
					entityId: "entity-1",
					positions: [
						{ start: 0, end: 3 }, // "テスト"は3文字
						{ start: 100, end: 110 }, // 範囲外
					],
				},
			];

			const result = generateEntityHighlightHTML("テスト", selectedEntities);

			const matches = result.match(/<a href="\/entity\/entity-1"/g);
			expect(matches).not.toBeNull();
			expect(matches).toHaveLength(1);
		});
	});
});
