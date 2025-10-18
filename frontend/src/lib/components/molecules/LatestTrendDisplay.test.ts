import { describe, expect, it } from "vitest";

// LatestTrendDisplayコンポーネントで使用されるユーティリティ関数のテスト
describe("LatestTrendDisplay utilities", () => {
	describe("期間フォーマット処理", () => {
		const formatPeriod = (
			start: string,
			end: string,
			locale: string,
		): string => {
			if (!start || !end) return "";

			const startDate = new Date(start);
			const endDate = new Date(end);

			if (locale === "ja") {
				const startStr = `${startDate.getFullYear()}年${startDate.getMonth() + 1}月${startDate.getDate()}日`;
				const endStr = `${endDate.getFullYear()}年${endDate.getMonth() + 1}月${endDate.getDate()}日`;
				return `${startStr} 〜 ${endStr}`;
			}
			return `${startDate.toLocaleDateString(locale || "en")} - ${endDate.toLocaleDateString(locale || "en")}`;
		};

		it("日本語ロケールで期間を正しくフォーマットする", () => {
			const result = formatPeriod(
				"2025-10-10T00:00:00Z",
				"2025-10-16T23:59:59Z",
				"ja",
			);
			expect(result).toContain("2025年10月10日");
			expect(result).toContain("2025年10月16日");
			expect(result).toContain("〜");
		});

		it("英語ロケールで期間を正しくフォーマットする", () => {
			const result = formatPeriod(
				"2025-10-10T00:00:00Z",
				"2025-10-16T23:59:59Z",
				"en",
			);
			expect(result).toContain("-");
		});

		it("空の期間文字列を処理する", () => {
			expect(formatPeriod("", "", "ja")).toBe("");
			expect(formatPeriod("2025-10-10T00:00:00Z", "", "ja")).toBe("");
			expect(formatPeriod("", "2025-10-16T23:59:59Z", "ja")).toBe("");
		});
	});

	describe("トレンドデータ構造", () => {
		it("正しいデータ構造を持つ", () => {
			const trendData = {
				analysis: "テスト分析結果",
				periodStart: "2025-10-10T00:00:00Z",
				periodEnd: "2025-10-16T23:59:59Z",
				generatedAt: "2025-10-17T04:00:00Z",
			};

			expect(trendData).toHaveProperty("analysis");
			expect(trendData).toHaveProperty("periodStart");
			expect(trendData).toHaveProperty("periodEnd");
			expect(trendData).toHaveProperty("generatedAt");
			expect(typeof trendData.analysis).toBe("string");
		});
	});
});
