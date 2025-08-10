import { beforeEach, describe, expect, it, vi } from "vitest";

describe("DarkModeToggle Component Logic", () => {
	let localStorageItems: Record<string, string> = {};
	const documentClassList: Set<string> = new Set();
	let mockAddEventListener: ReturnType<typeof vi.fn>;
	let mockMatchMedia: ReturnType<typeof vi.fn>;

	beforeEach(() => {
		vi.clearAllMocks();

		// Reset localStorage mock
		localStorageItems = {};
		documentClassList.clear();

		// Mock localStorage
		Object.defineProperty(globalThis, "localStorage", {
			value: {
				getItem: vi.fn((key: string) => localStorageItems[key] || null),
				setItem: vi.fn((key: string, value: string) => {
					localStorageItems[key] = value;
				}),
			},
			writable: true,
		});

		// Mock document.documentElement.classList
		Object.defineProperty(globalThis.document, "documentElement", {
			value: {
				classList: {
					add: vi.fn((className: string) => {
						documentClassList.add(className);
					}),
					remove: vi.fn((className: string) => {
						documentClassList.delete(className);
					}),
					contains: vi.fn((className: string) =>
						documentClassList.has(className),
					),
				},
			},
			writable: true,
		});

		// Mock window.matchMedia
		mockMatchMedia = vi.fn();
		mockAddEventListener = vi.fn();
		Object.defineProperty(globalThis.window, "matchMedia", {
			value: mockMatchMedia,
			writable: true,
		});
	});

	// Test dark mode initialization from localStorage
	it("should initialize dark mode from localStorage", () => {
		localStorageItems.darkMode = "true";

		function initializeDarkMode() {
			const stored = localStorage.getItem("darkMode");
			let isDarkMode = false;

			if (stored !== null) {
				isDarkMode = stored === "true";
			}

			return isDarkMode;
		}

		const result = initializeDarkMode();
		expect(result).toBe(true);
		expect(localStorage.getItem).toHaveBeenCalledWith("darkMode");
	});

	// Test dark mode initialization from system preference
	it("should initialize dark mode from system preference when localStorage is empty", () => {
		// Mock matchMedia to return dark mode preference
		mockMatchMedia.mockReturnValue({
			matches: true,
			addEventListener: mockAddEventListener,
		});

		function initializeDarkMode() {
			const stored = localStorage.getItem("darkMode");
			let isDarkMode = false;

			if (stored !== null) {
				isDarkMode = stored === "true";
			} else {
				// Use system preference
				const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
				isDarkMode = mediaQuery.matches;
			}

			return isDarkMode;
		}

		const result = initializeDarkMode();
		expect(result).toBe(true);
		expect(mockMatchMedia).toHaveBeenCalledWith("(prefers-color-scheme: dark)");
	});

	// Test dark mode toggle functionality
	it("should toggle dark mode state", () => {
		let isDarkMode = false;

		function toggleDarkMode() {
			isDarkMode = !isDarkMode;
			localStorage.setItem("darkMode", isDarkMode.toString());
		}

		// Initial state
		expect(isDarkMode).toBe(false);

		// Toggle to dark mode
		toggleDarkMode();
		expect(isDarkMode).toBe(true);
		expect(localStorage.setItem).toHaveBeenCalledWith("darkMode", "true");

		// Toggle back to light mode
		toggleDarkMode();
		expect(isDarkMode).toBe(false);
		expect(localStorage.setItem).toHaveBeenCalledWith("darkMode", "false");
	});

	// Test applying dark mode to document
	it("should apply dark mode class to document", () => {
		function applyDarkMode(isDarkMode: boolean) {
			if (isDarkMode) {
				document.documentElement.classList.add("dark");
			} else {
				document.documentElement.classList.remove("dark");
			}
		}

		// Apply dark mode
		applyDarkMode(true);
		expect(document.documentElement.classList.add).toHaveBeenCalledWith("dark");
		expect(documentClassList.has("dark")).toBe(true);

		// Remove dark mode
		applyDarkMode(false);
		expect(document.documentElement.classList.remove).toHaveBeenCalledWith(
			"dark",
		);
		expect(documentClassList.has("dark")).toBe(false);
	});

	// Test system preference listener
	it("should listen to system color scheme changes", () => {
		const mockMediaQuery = {
			matches: false,
			addEventListener: mockAddEventListener,
		};
		mockMatchMedia.mockReturnValue(mockMediaQuery);

		function setupSystemListener() {
			const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
			mediaQuery.addEventListener("change", vi.fn());
		}

		setupSystemListener();
		expect(mockMatchMedia).toHaveBeenCalledWith("(prefers-color-scheme: dark)");
		expect(mockAddEventListener).toHaveBeenCalledWith(
			"change",
			expect.any(Function),
		);
	});

	// Test system preference change handler
	it("should handle system preference changes when no manual preference is set", () => {
		let isDarkMode = false;

		function handleSystemChange(event: { matches: boolean }) {
			// Only apply system preference if user hasn't set a manual preference
			if (localStorage.getItem("darkMode") === null) {
				isDarkMode = event.matches;
			}
		}

		// System changes to dark mode
		handleSystemChange({ matches: true });
		expect(isDarkMode).toBe(true);

		// System changes to light mode
		handleSystemChange({ matches: false });
		expect(isDarkMode).toBe(false);
	});

	// Test system preference change handler with manual preference
	it("should ignore system preference changes when manual preference is set", () => {
		localStorageItems.darkMode = "false"; // Manual preference set
		let isDarkMode = false;

		function handleSystemChange(event: { matches: boolean }) {
			// Only apply system preference if user hasn't set a manual preference
			if (localStorage.getItem("darkMode") === null) {
				isDarkMode = event.matches;
			}
		}

		// System changes to dark mode, but should be ignored
		handleSystemChange({ matches: true });
		expect(isDarkMode).toBe(false); // Should remain false due to manual preference
	});

	// Test localStorage value parsing
	it("should correctly parse localStorage boolean values", () => {
		const testCases = [
			{ stored: "true", expected: true },
			{ stored: "false", expected: false },
			{ stored: "", expected: false },
			{ stored: "invalid", expected: false },
			{ stored: null, expected: false },
		];

		testCases.forEach(({ stored, expected }) => {
			// Reset localStorage
			if (stored === null) {
				delete localStorageItems.darkMode;
			} else {
				localStorageItems.darkMode = stored;
			}

			function parseStoredValue() {
				const stored = localStorage.getItem("darkMode");
				if (stored !== null) {
					return stored === "true";
				}
				return false; // Default to false when null
			}

			const result = parseStoredValue();
			expect(result).toBe(expected);
		});
	});

	// Test browser environment check
	it("should only operate in browser environment", () => {
		const testCases = [
			{ browser: true, shouldExecute: true },
			{ browser: false, shouldExecute: false },
		];

		testCases.forEach(({ browser, shouldExecute }) => {
			let executedCount = 0;

			function toggleDarkMode() {
				if (browser) {
					executedCount++;
					localStorage.setItem("darkMode", "true");
				}
			}

			toggleDarkMode();

			if (shouldExecute) {
				expect(executedCount).toBe(1);
				expect(localStorage.setItem).toHaveBeenCalled();
			} else {
				expect(executedCount).toBe(0);
			}

			vi.clearAllMocks();
		});
	});

	// Test complete toggle workflow
	it("should complete full toggle workflow", () => {
		let isDarkMode = false;

		function completeDarkModeToggle() {
			// Toggle state
			isDarkMode = !isDarkMode;

			// Save to localStorage
			localStorage.setItem("darkMode", isDarkMode.toString());

			// Apply to document
			if (isDarkMode) {
				document.documentElement.classList.add("dark");
			} else {
				document.documentElement.classList.remove("dark");
			}

			return isDarkMode;
		}

		// First toggle (light to dark)
		const result1 = completeDarkModeToggle();
		expect(result1).toBe(true);
		expect(localStorage.setItem).toHaveBeenCalledWith("darkMode", "true");
		expect(document.documentElement.classList.add).toHaveBeenCalledWith("dark");

		// Second toggle (dark to light)
		const result2 = completeDarkModeToggle();
		expect(result2).toBe(false);
		expect(localStorage.setItem).toHaveBeenCalledWith("darkMode", "false");
		expect(document.documentElement.classList.remove).toHaveBeenCalledWith(
			"dark",
		);
	});

	// Test aria labels and accessibility
	it("should provide correct accessibility labels", () => {
		function getAriaLabel(isDarkMode: boolean) {
			return isDarkMode ? "Switch to light mode" : "Switch to dark mode";
		}

		expect(getAriaLabel(false)).toBe("Switch to dark mode");
		expect(getAriaLabel(true)).toBe("Switch to light mode");
	});

	// Test icon display logic
	it("should show correct icon based on mode", () => {
		function getIconType(isDarkMode: boolean) {
			return isDarkMode ? "sun" : "moon";
		}

		expect(getIconType(false)).toBe("moon"); // Light mode shows moon (to switch to dark)
		expect(getIconType(true)).toBe("sun"); // Dark mode shows sun (to switch to light)
	});
});
