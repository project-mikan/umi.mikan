import { describe, expect, it, vi } from "vitest";
import { load } from "./+page.server";
import * as diaryApi from "$lib/server/diary-api";
import type { PageServerLoad } from "./$types";
import { create } from "@bufbuild/protobuf";
import {
	YMDSchema,
	DiaryEntrySchema,
	GetDiaryCountResponseSchema,
	GetDiaryEntryResponseSchema,
} from "$lib/grpc/diary/diary_pb";

// Mock the diary API module
vi.mock("$lib/server/diary-api", () => ({
	getDiaryEntry: vi.fn(),
	getDiaryCount: vi.fn(),
	createYMD: vi.fn(),
	createDiaryEntry: vi.fn(),
	updateDiaryEntry: vi.fn(),
}));

describe("+page.server load function", () => {
	it("should load diary data with diary count", async () => {
		// Mock the current date to be predictable
		const mockDate = new Date("2024-01-15T12:00:00Z");
		vi.setSystemTime(mockDate);

		// Setup mocks
		const mockYMD = create(YMDSchema, { year: 2024, month: 1, day: 15 });
		const mockEntry = create(DiaryEntrySchema, {
			id: "test-entry-1",
			content: "Test diary entry",
			date: mockYMD,
		});

		vi.mocked(diaryApi.createYMD).mockReturnValue(mockYMD);
		vi.mocked(diaryApi.getDiaryEntry).mockResolvedValue(
			create(GetDiaryEntryResponseSchema, { entry: mockEntry }),
		);
		vi.mocked(diaryApi.getDiaryCount).mockResolvedValue(
			create(GetDiaryCountResponseSchema, { count: 42 }),
		);

		// Mock cookies
		const mockCookies = new Map([["accessToken", "test-token"]]);
		const mockRequestEvent = {
			cookies: {
				get: (name: string) => mockCookies.get(name),
			},
		} as Parameters<PageServerLoad>[0];

		// Call the load function
		const result = await load(mockRequestEvent);

		// Verify the result includes diary count
		expect(result?.diaryCount).toBe(42);
		expect(result?.today).toBeDefined();
		expect(result?.yesterday).toBeDefined();
		expect(result?.dayBeforeYesterday).toBeDefined();

		// Verify API calls were made
		expect(diaryApi.getDiaryCount).toHaveBeenCalledWith({
			accessToken: "test-token",
		});
		expect(diaryApi.getDiaryEntry).toHaveBeenCalledTimes(3);
	});

	it("should return diary count as 0 when API call fails", async () => {
		const mockDate = new Date("2024-01-15T12:00:00Z");
		vi.setSystemTime(mockDate);

		const mockYMD = create(YMDSchema, { year: 2024, month: 1, day: 15 });

		vi.mocked(diaryApi.createYMD).mockReturnValue(mockYMD);
		vi.mocked(diaryApi.getDiaryEntry).mockRejectedValue(new Error("API Error"));
		vi.mocked(diaryApi.getDiaryCount).mockRejectedValue(
			new Error("Count API Error"),
		);

		const mockCookies = new Map([["accessToken", "test-token"]]);
		const mockRequestEvent = {
			cookies: {
				get: (name: string) => mockCookies.get(name),
			},
		} as Parameters<PageServerLoad>[0];

		const result = await load(mockRequestEvent);

		// Should return 0 count when API fails
		expect(result?.diaryCount).toBe(0);
		expect(result?.today.entry).toBeNull();
		expect(result?.yesterday.entry).toBeNull();
		expect(result?.dayBeforeYesterday.entry).toBeNull();
	});

	it("should handle partial API failures gracefully", async () => {
		const mockDate = new Date("2024-01-15T12:00:00Z");
		vi.setSystemTime(mockDate);

		const mockYMD = create(YMDSchema, { year: 2024, month: 1, day: 15 });
		const mockEntry = create(DiaryEntrySchema, {
			id: "test-entry-1",
			content: "Test diary entry",
			date: mockYMD,
		});

		vi.mocked(diaryApi.createYMD).mockReturnValue(mockYMD);

		// Mock partial failures: getDiaryEntry succeeds, getDiaryCount fails
		vi.mocked(diaryApi.getDiaryEntry).mockResolvedValue(
			create(GetDiaryEntryResponseSchema, { entry: mockEntry }),
		);
		vi.mocked(diaryApi.getDiaryCount).mockRejectedValue(
			new Error("Count API Error"),
		);

		const mockCookies = new Map([["accessToken", "test-token"]]);
		const mockRequestEvent = {
			cookies: {
				get: (name: string) => mockCookies.get(name),
			},
		} as Parameters<PageServerLoad>[0];

		const result = await load(mockRequestEvent);

		// Should handle mixed success/failure
		expect(result?.diaryCount).toBe(0); // Failed API call
		expect(result?.today.entry).toEqual(mockEntry); // Successful API call
		expect(result?.yesterday.entry).toEqual(mockEntry);
		expect(result?.dayBeforeYesterday.entry).toEqual(mockEntry);
	});

	it("should redirect when no access token", async () => {
		const mockCookies = new Map(); // No access token
		const mockRequestEvent = {
			cookies: {
				get: (name: string) => mockCookies.get(name),
			},
		} as Parameters<PageServerLoad>[0];

		// Should throw redirect error
		await expect(load(mockRequestEvent)).rejects.toThrow();
	});
});

describe("Date creation utilities", () => {
	it("should create correct dates for load function", () => {
		// Test the date logic used in the load function
		const testDate = new Date("2024-03-15T12:00:00Z");

		const today = {
			year: testDate.getFullYear(),
			month: testDate.getMonth() + 1,
			day: testDate.getDate(),
		};

		const yesterday = new Date(testDate);
		yesterday.setDate(yesterday.getDate() - 1);
		const yesterdayYMD = {
			year: yesterday.getFullYear(),
			month: yesterday.getMonth() + 1,
			day: yesterday.getDate(),
		};

		const dayBeforeYesterday = new Date(testDate);
		dayBeforeYesterday.setDate(dayBeforeYesterday.getDate() - 2);
		const dayBeforeYesterdayYMD = {
			year: dayBeforeYesterday.getFullYear(),
			month: dayBeforeYesterday.getMonth() + 1,
			day: dayBeforeYesterday.getDate(),
		};

		expect(today).toEqual({ year: 2024, month: 3, day: 15 });
		expect(yesterdayYMD).toEqual({ year: 2024, month: 3, day: 14 });
		expect(dayBeforeYesterdayYMD).toEqual({ year: 2024, month: 3, day: 13 });
	});
});
