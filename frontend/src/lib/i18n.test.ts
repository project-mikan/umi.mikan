import { beforeEach, describe, expect, it, vi } from "vitest";

// Mock browser environment
vi.mock("$app/environment", () => ({
	browser: true,
}));

// Mock svelte-i18n
const mockInit = vi.fn();
const mockRegister = vi.fn();
const mockLocale = { subscribe: vi.fn() };
const mockWaitLocale = vi.fn();

vi.mock("svelte-i18n", () => ({
	init: mockInit,
	register: mockRegister,
	locale: mockLocale,
	waitLocale: mockWaitLocale,
}));

// Mock locale imports
vi.mock("../locales/en.json", () => ({ default: {} }));
vi.mock("../locales/ja.json", () => ({ default: {} }));

describe("i18n", () => {
	beforeEach(() => {
		vi.clearAllMocks();

		// Mock navigator.language
		Object.defineProperty(window.navigator, "language", {
			writable: true,
			value: "en-US",
		});
	});

	it("should register locales correctly", async () => {
		await import("./i18n.ts");

		expect(mockRegister).toHaveBeenCalledWith("en", expect.any(Function));
		expect(mockRegister).toHaveBeenCalledWith("ja", expect.any(Function));
	});

	it("should initialize with English locale for non-Japanese browsers", async () => {
		Object.defineProperty(window.navigator, "language", {
			writable: true,
			value: "en-US",
		});

		// Clear module cache to re-import with new navigator.language
		vi.resetModules();

		await import("./i18n.ts");

		expect(mockInit).toHaveBeenCalledWith({
			fallbackLocale: "en",
			initialLocale: "en",
		});
	});

	it("should initialize with Japanese locale for Japanese browsers", async () => {
		Object.defineProperty(window.navigator, "language", {
			writable: true,
			value: "ja-JP",
		});

		// Clear module cache to re-import with new navigator.language
		vi.resetModules();

		await import("./i18n.ts");

		expect(mockInit).toHaveBeenCalledWith({
			fallbackLocale: "en",
			initialLocale: "ja",
		});
	});

	it("should export locale and waitLocale", async () => {
		const i18nModule = await import("./i18n.ts");

		expect(i18nModule.locale).toBeDefined();
		expect(i18nModule.waitLocale).toBeDefined();
	});
});
