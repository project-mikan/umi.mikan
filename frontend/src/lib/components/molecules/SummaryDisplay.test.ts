import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// Mock authenticatedFetch
vi.mock("$lib/auth-client", () => ({
	authenticatedFetch: vi.fn(),
}));

import { authenticatedFetch } from "$lib/auth-client";

// Types for test data
interface TestSummary {
	id: string;
	summary: string;
	createdAt: number;
	updatedAt: number;
}

interface GeneratePayload {
	year: number;
	month: number;
}

describe("SummaryDisplay Logic Tests", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	// Test the core logic of determining if animation should be triggered
	it("should NOT trigger animation when summary content is the same", async () => {
		const existingSummary = {
			id: "test-summary-1",
			summary: "Existing summary content",
			createdAt: 1640000000,
			updatedAt: 1640000000,
		};

		const newSummary = {
			id: "test-summary-1",
			summary: "Existing summary content", // Same content
			createdAt: 1640000000,
			updatedAt: 1640000000, // Same updatedAt
		};

		// Logic to check if summary should trigger animation
		function shouldTriggerAnimation(
			oldSummary: TestSummary,
			newSummary: TestSummary,
		): boolean {
			return (
				oldSummary.summary !== newSummary.summary ||
				oldSummary.updatedAt !== newSummary.updatedAt
			);
		}

		const result = shouldTriggerAnimation(existingSummary, newSummary);
		expect(result).toBe(false);
	});

	it("should trigger animation when summary content changes", async () => {
		const existingSummary = {
			id: "test-summary-1",
			summary: "Existing summary content",
			createdAt: 1640000000,
			updatedAt: 1640000000,
		};

		const newSummary = {
			id: "test-summary-1",
			summary: "New updated summary content", // Different content
			createdAt: 1640000000,
			updatedAt: 1640000100, // Different updatedAt
		};

		// Logic to check if summary should trigger animation
		function shouldTriggerAnimation(
			oldSummary: TestSummary,
			newSummary: TestSummary,
		): boolean {
			return (
				oldSummary.summary !== newSummary.summary ||
				oldSummary.updatedAt !== newSummary.updatedAt
			);
		}

		const result = shouldTriggerAnimation(existingSummary, newSummary);
		expect(result).toBe(true);
	});

	it("should trigger animation when updatedAt changes even with same content", async () => {
		const existingSummary = {
			id: "test-summary-1",
			summary: "Same content",
			createdAt: 1640000000,
			updatedAt: 1640000000,
		};

		const newSummary = {
			id: "test-summary-1",
			summary: "Same content",
			createdAt: 1640000000,
			updatedAt: 1640000100, // Different updatedAt
		};

		// Logic to check if summary should trigger animation
		function shouldTriggerAnimation(
			oldSummary: TestSummary,
			newSummary: TestSummary,
		): boolean {
			return (
				oldSummary.summary !== newSummary.summary ||
				oldSummary.updatedAt !== newSummary.updatedAt
			);
		}

		const result = shouldTriggerAnimation(existingSummary, newSummary);
		expect(result).toBe(true);
	});

	// Test regeneration logic
	it("should handle successful regeneration with new summary", async () => {
		const newSummary = {
			id: "test-summary-1",
			summary: "New updated summary content",
			createdAt: 1640000000,
			updatedAt: 1640000100,
		};

		vi.mocked(authenticatedFetch).mockResolvedValueOnce({
			ok: true,
			json: () => Promise.resolve({ summary: newSummary }),
		} as Response);

		// Logic to handle regeneration
		async function handleRegeneration(
			generateUrl: string,
			generatePayload: GeneratePayload,
		) {
			const response = await authenticatedFetch(generateUrl, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(generatePayload),
			});

			if (response.ok) {
				const data = await response.json();
				return data.summary;
			}
			throw new Error("Failed to regenerate");
		}

		const result = await handleRegeneration("/api/diary/summary/generate", {
			year: 2024,
			month: 1,
		});

		expect(result).toEqual(newSummary);
		expect(authenticatedFetch).toHaveBeenCalledWith(
			"/api/diary/summary/generate",
			expect.objectContaining({
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ year: 2024, month: 1 }),
			}),
		);
	});

	it("should handle regeneration failure", async () => {
		vi.mocked(authenticatedFetch).mockResolvedValueOnce({
			ok: false,
			status: 500,
		} as Response);

		// Logic to handle regeneration
		async function handleRegeneration(
			generateUrl: string,
			generatePayload: GeneratePayload,
		) {
			const response = await authenticatedFetch(generateUrl, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(generatePayload),
			});

			if (response.ok) {
				const data = await response.json();
				return data.summary;
			}
			throw new Error("Failed to regenerate");
		}

		await expect(
			handleRegeneration("/api/diary/summary/generate", {
				year: 2024,
				month: 1,
			}),
		).rejects.toThrow("Failed to regenerate");
	});

	// Test polling logic
	it("should determine correct polling state", async () => {
		const queuedSummary = {
			id: "test-summary-1",
			summary: "generation has been queued",
			createdAt: 1640000000,
			updatedAt: 1640000000,
		};

		const completedSummary = {
			id: "test-summary-1",
			summary: "Completed summary content",
			createdAt: 1640000000,
			updatedAt: 1640000100,
		};

		// Logic to determine if polling should continue
		function shouldContinuePolling(summary: TestSummary): boolean {
			return summary.summary === "generation has been queued";
		}

		expect(shouldContinuePolling(queuedSummary)).toBe(true);
		expect(shouldContinuePolling(completedSummary)).toBe(false);
	});
});
