import { describe, expect, it } from "vitest";
import ja from "../locales/ja.json";
import en from "../locales/en.json";

describe("Diary Count Internationalization", () => {
	describe("Japanese locale", () => {
		it("should have totalCount translation key", () => {
			expect(ja.diary.totalCount).toBeDefined();
			expect(ja.diary.totalCount).toBe("これまでに{count}日分の日記を書きました");
		});

		it("should contain count placeholder", () => {
			expect(ja.diary.totalCount).toContain("{count}");
		});

		it("should be properly formatted for interpolation", () => {
			const translation = ja.diary.totalCount;
			const testCount = 5;
			const result = translation.replace("{count}", testCount.toString());
			expect(result).toBe("これまでに5日分の日記を書きました");
		});
	});

	describe("English locale", () => {
		it("should have totalCount translation key", () => {
			expect(en.diary.totalCount).toBeDefined();
			expect(en.diary.totalCount).toBe("You have written {count} diary entries so far");
		});

		it("should contain count placeholder", () => {
			expect(en.diary.totalCount).toContain("{count}");
		});

		it("should be properly formatted for interpolation", () => {
			const translation = en.diary.totalCount;
			const testCount = 5;
			const result = translation.replace("{count}", testCount.toString());
			expect(result).toBe("You have written 5 diary entries so far");
		});
	});

	describe("Translation consistency", () => {
		it("should have same structure in both locales", () => {
			expect(typeof ja.diary.totalCount).toBe("string");
			expect(typeof en.diary.totalCount).toBe("string");
		});

		it("should both contain the count placeholder", () => {
			expect(ja.diary.totalCount.includes("{count}")).toBe(true);
			expect(en.diary.totalCount.includes("{count}")).toBe(true);
		});

		it("should not be empty strings", () => {
			expect(ja.diary.totalCount.trim()).not.toBe("");
			expect(en.diary.totalCount.trim()).not.toBe("");
		});
	});

	describe("Edge cases", () => {
		it("should handle zero count", () => {
			const jaResult = ja.diary.totalCount.replace("{count}", "0");
			const enResult = en.diary.totalCount.replace("{count}", "0");

			expect(jaResult).toBe("これまでに0日分の日記を書きました");
			expect(enResult).toBe("You have written 0 diary entries so far");
		});

		it("should handle large count", () => {
			const largeCount = "999";
			const jaResult = ja.diary.totalCount.replace("{count}", largeCount);
			const enResult = en.diary.totalCount.replace("{count}", largeCount);

			expect(jaResult).toBe("これまでに999日分の日記を書きました");
			expect(enResult).toBe("You have written 999 diary entries so far");
		});

		it("should handle single count appropriately", () => {
			const singleCount = "1";
			const jaResult = ja.diary.totalCount.replace("{count}", singleCount);
			const enResult = en.diary.totalCount.replace("{count}", singleCount);

			expect(jaResult).toBe("これまでに1日分の日記を書きました");
			expect(enResult).toBe("You have written 1 diary entries so far");
			// Note: English could use better pluralization but this is current implementation
		});
	});

	describe("Locale completeness for diary feature", () => {
		it("should have all required diary keys in both locales", () => {
			const requiredKeys = ["title", "totalCount", "today", "yesterday"];
			
			for (const key of requiredKeys) {
				expect(ja.diary[key as keyof typeof ja.diary]).toBeDefined();
				expect(en.diary[key as keyof typeof en.diary]).toBeDefined();
			}
		});

		it("should not have null or undefined values", () => {
			expect(ja.diary.totalCount).not.toBeNull();
			expect(ja.diary.totalCount).not.toBeUndefined();
			expect(en.diary.totalCount).not.toBeNull();
			expect(en.diary.totalCount).not.toBeUndefined();
		});
	});
});