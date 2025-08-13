import { describe, expect, it } from "vitest";
import {
	formatDiaryCountMessage,
	formatDateStr,
	getMonthlyUrl,
} from "./diary-count-utils";

describe("Diary Count Utilities", () => {
	describe("formatDiaryCountMessage", () => {
		it("should format English message correctly", () => {
			expect(formatDiaryCountMessage(5, "en")).toBe(
				"You have written 5 diary entries so far",
			);
			expect(formatDiaryCountMessage(1, "en")).toBe(
				"You have written 1 diary entries so far",
			);
			expect(formatDiaryCountMessage(0, "en")).toBe(
				"You have written 0 diary entries so far",
			);
		});

		it("should format Japanese message correctly", () => {
			expect(formatDiaryCountMessage(5, "ja")).toBe(
				"これまでに5日分の日記を書きました",
			);
			expect(formatDiaryCountMessage(1, "ja")).toBe(
				"これまでに1日分の日記を書きました",
			);
			expect(formatDiaryCountMessage(0, "ja")).toBe(
				"これまでに0日分の日記を書きました",
			);
		});

		it("should default to English when locale not specified", () => {
			expect(formatDiaryCountMessage(10)).toBe(
				"You have written 10 diary entries so far",
			);
		});

		it("should handle large numbers", () => {
			expect(formatDiaryCountMessage(999, "en")).toBe(
				"You have written 999 diary entries so far",
			);
			expect(formatDiaryCountMessage(1000, "ja")).toBe(
				"これまでに1000日分の日記を書きました",
			);
		});
	});

	describe("formatDateStr", () => {
		it("should format dates with proper padding", () => {
			expect(formatDateStr({ year: 2024, month: 1, day: 5 })).toBe(
				"2024-01-05",
			);
			expect(formatDateStr({ year: 2024, month: 12, day: 25 })).toBe(
				"2024-12-25",
			);
			expect(formatDateStr({ year: 2024, month: 10, day: 1 })).toBe(
				"2024-10-01",
			);
		});

		it("should handle edge cases", () => {
			expect(formatDateStr({ year: 2000, month: 1, day: 1 })).toBe(
				"2000-01-01",
			);
			expect(formatDateStr({ year: 9999, month: 12, day: 31 })).toBe(
				"9999-12-31",
			);
		});

		it("should pad single digit months and days", () => {
			expect(formatDateStr({ year: 2024, month: 3, day: 7 })).toBe(
				"2024-03-07",
			);
			expect(formatDateStr({ year: 2024, month: 11, day: 9 })).toBe(
				"2024-11-09",
			);
		});
	});

	describe("getMonthlyUrl", () => {
		it("should generate correct URLs for given dates", () => {
			const janDate = new Date("2024-01-15T12:00:00Z");
			expect(getMonthlyUrl(janDate)).toBe("/monthly/2024/1");

			const decDate = new Date("2024-12-15T12:00:00Z");
			expect(getMonthlyUrl(decDate)).toBe("/monthly/2024/12");

			const marchDate = new Date("2024-03-15T12:00:00Z");
			expect(getMonthlyUrl(marchDate)).toBe("/monthly/2024/3");
		});

		it("should use current date when no date provided", () => {
			// Test that it returns a valid URL pattern without hardcoding current date
			const url = getMonthlyUrl();
			expect(url).toMatch(/^\/monthly\/\d{4}\/\d{1,2}$/);
		});

		it("should handle edge case months", () => {
			const januaryDate = new Date("2024-01-01T12:00:00Z");
			expect(getMonthlyUrl(januaryDate)).toBe("/monthly/2024/1");

			const decemberDate = new Date("2024-12-31T12:00:00Z");
			expect(getMonthlyUrl(decemberDate)).toBe("/monthly/2024/12");
		});
	});

	describe("Integration tests", () => {
		it("should work together to create diary display data", () => {
			const mockData = {
				diaryCount: 25,
				today: {
					date: { year: 2024, month: 3, day: 15 },
				},
			};

			const countMessage = formatDiaryCountMessage(mockData.diaryCount, "en");
			const dateStr = formatDateStr(mockData.today.date);
			const monthlyUrl = getMonthlyUrl(new Date("2024-03-15"));

			expect(countMessage).toBe("You have written 25 diary entries so far");
			expect(dateStr).toBe("2024-03-15");
			expect(monthlyUrl).toBe("/monthly/2024/3");
		});
	});
});
