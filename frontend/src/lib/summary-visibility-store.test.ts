import { describe, it, expect, beforeEach, vi, afterEach } from "vitest";
import { get } from "svelte/store";
import { summaryVisibility } from "./summary-visibility-store";

// localStorage のモック
const localStorageMock = {
	getItem: vi.fn(),
	setItem: vi.fn(),
	removeItem: vi.fn(),
	clear: vi.fn(),
};

// globalThis オブジェクトに localStorage を設定
Object.defineProperty(globalThis, "localStorage", {
	value: localStorageMock,
	writable: true,
});

// browser のモック
vi.mock("$app/environment", () => ({
	browser: true,
}));

describe("summaryVisibility store", () => {
	beforeEach(() => {
		vi.clearAllMocks();
		// デフォルト状態にリセット
		localStorageMock.getItem.mockReturnValue(null);
		// ストアをリセット
		(
			summaryVisibility as typeof summaryVisibility & { _reset: () => void }
		)._reset();
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	it("should initialize with default values when localStorage is empty", () => {
		summaryVisibility.init();
		const state = get(summaryVisibility);

		expect(state.daily).toBe(true);
		expect(state.monthly).toBe(true);
		expect(state.latestTrend).toBe(true);
	});

	it("should only initialize once", () => {
		localStorageMock.getItem.mockReturnValue(
			JSON.stringify({ daily: false, monthly: false }),
		);

		summaryVisibility.init();
		let state = get(summaryVisibility);
		expect(state.daily).toBe(false);
		expect(state.monthly).toBe(false);

		// 2回目の初期化では状態は変わらない
		localStorageMock.getItem.mockReturnValue(
			JSON.stringify({ daily: true, monthly: true }),
		);
		summaryVisibility.init();
		state = get(summaryVisibility);
		expect(state.daily).toBe(false); // 変わらない
		expect(state.monthly).toBe(false); // 変わらない
	});

	it("should load values from localStorage when available", () => {
		const storedData = { daily: false, monthly: true };
		localStorageMock.getItem.mockReturnValue(JSON.stringify(storedData));

		summaryVisibility.init();
		const state = get(summaryVisibility);

		expect(localStorageMock.getItem).toHaveBeenCalledWith("summary-visibility");
		expect(state.daily).toBe(false);
		expect(state.monthly).toBe(true);
	});

	it("should merge stored values with defaults", () => {
		const storedData = { daily: false }; // monthly は含まれていない
		localStorageMock.getItem.mockReturnValue(JSON.stringify(storedData));

		summaryVisibility.init();
		const state = get(summaryVisibility);

		expect(state.daily).toBe(false);
		expect(state.monthly).toBe(true); // デフォルト値が使用される
	});

	it("should handle invalid localStorage data gracefully", () => {
		localStorageMock.getItem.mockReturnValue("invalid json");

		summaryVisibility.init();
		const state = get(summaryVisibility);

		expect(state.daily).toBe(true);
		expect(state.monthly).toBe(true);
	});

	it("should toggle daily visibility and save to localStorage", () => {
		summaryVisibility.init();

		// 初期状態をチェック
		let state = get(summaryVisibility);
		expect(state.daily).toBe(true);

		// daily をトグル
		summaryVisibility.toggleDaily();
		state = get(summaryVisibility);

		expect(state.daily).toBe(false);
		expect(state.monthly).toBe(true); // monthly は変わらない
		expect(localStorageMock.setItem).toHaveBeenCalledWith(
			"summary-visibility",
			JSON.stringify({ daily: false, monthly: true, latestTrend: true }),
		);

		// もう一度トグル
		summaryVisibility.toggleDaily();
		state = get(summaryVisibility);

		expect(state.daily).toBe(true);
		expect(localStorageMock.setItem).toHaveBeenCalledWith(
			"summary-visibility",
			JSON.stringify({ daily: true, monthly: true, latestTrend: true }),
		);
	});

	it("should toggle monthly visibility and save to localStorage", () => {
		summaryVisibility.init();

		// 初期状態をチェック
		let state = get(summaryVisibility);
		expect(state.monthly).toBe(true);

		// monthly をトグル
		summaryVisibility.toggleMonthly();
		state = get(summaryVisibility);

		expect(state.monthly).toBe(false);
		expect(state.daily).toBe(true); // daily は変わらない
		expect(localStorageMock.setItem).toHaveBeenCalledWith(
			"summary-visibility",
			JSON.stringify({ daily: true, monthly: false, latestTrend: true }),
		);

		// もう一度トグル
		summaryVisibility.toggleMonthly();
		state = get(summaryVisibility);

		expect(state.monthly).toBe(true);
		expect(localStorageMock.setItem).toHaveBeenCalledWith(
			"summary-visibility",
			JSON.stringify({ daily: true, monthly: true, latestTrend: true }),
		);
	});

	it("should handle localStorage errors gracefully", () => {
		// setItem でエラーが発生する場合
		localStorageMock.setItem.mockImplementation(() => {
			throw new Error("Storage quota exceeded");
		});

		summaryVisibility.init();

		// エラーが発生してもトグルは機能する
		expect(() => summaryVisibility.toggleDaily()).not.toThrow();

		const state = get(summaryVisibility);
		expect(state.daily).toBe(false); // ストアの状態は更新される
	});
});
