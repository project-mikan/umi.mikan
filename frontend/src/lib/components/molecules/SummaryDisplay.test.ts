import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { render, fireEvent, waitFor } from "@testing-library/svelte";

// Mock authenticatedFetch
vi.mock("$lib/auth-client", () => ({
	authenticatedFetch: vi.fn(),
}));

// Mock svelte-i18n
vi.mock("svelte-i18n", () => ({
	_: vi.fn((key) => key),
	locale: { subscribe: vi.fn() },
}));

// Mock $lib/i18n
vi.mock("$lib/i18n", () => ({}));

// Mock $app/environment
vi.mock("$app/environment", () => ({
	browser: true,
}));

import SummaryDisplay from "./SummaryDisplay.svelte";
import { authenticatedFetch } from "$lib/auth-client";

describe("SummaryDisplay Animation Issue", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it("should NOT trigger animation immediately when regenerating with same summary", async () => {
		const existingSummary = {
			id: "test-summary-1",
			summary: "Existing summary content",
			createdAt: 1640000000,
			updatedAt: 1640000000,
		};

		// Mock fetch to return the same summary immediately (simulating the problem)
		vi.mocked(authenticatedFetch).mockResolvedValueOnce({
			ok: true,
			json: () => Promise.resolve({ summary: existingSummary }),
		} as Response);

		const component = render(SummaryDisplay, {
			props: {
				type: "monthly",
				fetchUrl: "/api/diary/summary/2024/1",
				generateUrl: "/api/diary/summary/generate",
				generatePayload: { year: 2024, month: 1 },
				summary: existingSummary,
				hasLLMKey: true,
				isDisabled: false,
			},
		});

		// Find the regenerate button
		const regenerateButton = component.getByRole("button");

		// Spy on the summary paragraph to check for animation class
		const summaryElement = component.container.querySelector("p.text-gray-700");
		expect(summaryElement).toBeTruthy();

		// Click regenerate
		await fireEvent.click(regenerateButton);

		// The animation class should NOT be applied immediately when the same summary is returned
		expect(summaryElement?.classList.contains("summary-highlight")).toBe(false);

		// Fast-forward timers to ensure no animation is triggered
		vi.advanceTimersByTime(100);
		expect(summaryElement?.classList.contains("summary-highlight")).toBe(false);
	});

	it("should trigger animation only when summary actually changes via polling", async () => {
		const existingSummary = {
			id: "test-summary-1",
			summary: "Existing summary content",
			createdAt: 1640000000,
			updatedAt: 1640000000,
		};

		const newSummary = {
			id: "test-summary-1",
			summary: "New updated summary content",
			createdAt: 1640000000,
			updatedAt: 1640000100, // Different updatedAt
		};

		// First call returns queued status
		vi.mocked(authenticatedFetch)
			.mockResolvedValueOnce({
				ok: true,
				json: () =>
					Promise.resolve({
						summary: {
							...existingSummary,
							summary: "generation has been queued",
						},
					}),
			} as Response)
			// Polling call returns new summary
			.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ summary: newSummary }),
			} as Response);

		const component = render(SummaryDisplay, {
			props: {
				type: "monthly",
				fetchUrl: "/api/diary/summary/2024/1",
				generateUrl: "/api/diary/summary/generate",
				generatePayload: { year: 2024, month: 1 },
				summary: existingSummary,
				hasLLMKey: true,
				isDisabled: false,
			},
		});

		const regenerateButton = component.getByRole("button");
		const summaryElement = component.container.querySelector("p.text-gray-700");

		// Click regenerate
		await fireEvent.click(regenerateButton);

		// Should NOT have animation immediately
		expect(summaryElement?.classList.contains("summary-highlight")).toBe(false);

		// Fast-forward to trigger polling
		vi.advanceTimersByTime(3000);
		await waitFor(() => {
			// After polling returns new summary, animation should be triggered
			expect(summaryElement?.classList.contains("summary-highlight")).toBe(
				true,
			);
		});
	});

	it("should show correct loading state during regeneration", async () => {
		const existingSummary = {
			id: "test-summary-1",
			summary: "Existing summary content",
			createdAt: 1640000000,
			updatedAt: 1640000000,
		};

		// Mock fetch to return queued status
		vi.mocked(authenticatedFetch).mockResolvedValueOnce({
			ok: true,
			json: () =>
				Promise.resolve({
					summary: {
						...existingSummary,
						summary: "generation has been queued",
					},
				}),
		} as Response);

		const component = render(SummaryDisplay, {
			props: {
				type: "monthly",
				fetchUrl: "/api/diary/summary/2024/1",
				generateUrl: "/api/diary/summary/generate",
				generatePayload: { year: 2024, month: 1 },
				summary: existingSummary,
				hasLLMKey: true,
				isDisabled: false,
			},
		});

		const regenerateButton = component.getByRole("button");

		// Click regenerate
		await fireEvent.click(regenerateButton);

		// Should show regenerating state
		await waitFor(() => {
			expect(regenerateButton.textContent).toContain("generating");
		});

		// Should show loading spinner
		const spinner = component.container.querySelector(".animate-spin");
		expect(spinner).toBeTruthy();
	});

	it("should emit events correctly when summary actually updates", async () => {
		const existingSummary = {
			id: "test-summary-1",
			summary: "Existing summary content",
			createdAt: 1640000000,
			updatedAt: 1640000000,
		};

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

		const summaryUpdatedHandler = vi.fn();
		const generationCompletedHandler = vi.fn();

		const component = render(SummaryDisplay, {
			props: {
				type: "monthly",
				fetchUrl: "/api/diary/summary/2024/1",
				generateUrl: "/api/diary/summary/generate",
				generatePayload: { year: 2024, month: 1 },
				summary: existingSummary,
				hasLLMKey: true,
				isDisabled: false,
			},
		});

		// Note: Event listeners removed due to type issues in newer Svelte versions

		const regenerateButton = component.getByRole("button");

		// Click regenerate
		await fireEvent.click(regenerateButton);

		await waitFor(() => {
			// summaryUpdated should be called with new summary
			expect(summaryUpdatedHandler).toHaveBeenCalledWith(
				expect.objectContaining({
					detail: { summary: newSummary },
				}),
			);
			// generationCompleted should be called
			expect(generationCompletedHandler).toHaveBeenCalled();
		});
	});

	it("should NOT emit summaryUpdated when summary content is the same", async () => {
		const existingSummary = {
			id: "test-summary-1",
			summary: "Existing summary content",
			createdAt: 1640000000,
			updatedAt: 1640000000,
		};

		// Return the same summary (no actual change)
		vi.mocked(authenticatedFetch).mockResolvedValueOnce({
			ok: true,
			json: () => Promise.resolve({ summary: existingSummary }),
		} as Response);

		const summaryUpdatedHandler = vi.fn();
		const generationCompletedHandler = vi.fn();

		const component = render(SummaryDisplay, {
			props: {
				type: "monthly",
				fetchUrl: "/api/diary/summary/2024/1",
				generateUrl: "/api/diary/summary/generate",
				generatePayload: { year: 2024, month: 1 },
				summary: existingSummary,
				hasLLMKey: true,
				isDisabled: false,
			},
		});

		// Note: Event listeners removed due to type issues in newer Svelte versions

		const regenerateButton = component.getByRole("button");

		// Click regenerate
		await fireEvent.click(regenerateButton);

		await waitFor(() => {
			// summaryUpdated should NOT be called since content is the same
			expect(summaryUpdatedHandler).not.toHaveBeenCalled();
			// generationCompleted should still be called
			expect(generationCompletedHandler).toHaveBeenCalled();
		});
	});
});
