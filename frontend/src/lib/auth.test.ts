import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";
import { isAuthenticated, user, logout } from "./auth";

describe("Auth Store", () => {
	beforeEach(() => {
		isAuthenticated.set(false);
		user.set(null);
	});

	it("should initialize with correct default values", () => {
		expect(get(isAuthenticated)).toBe(false);
		expect(get(user)).toBeNull();
	});

	it("should update isAuthenticated store", () => {
		isAuthenticated.set(true);
		expect(get(isAuthenticated)).toBe(true);
	});

	it("should update user store", () => {
		const mockUser = { id: "1", email: "test@example.com", name: "Test User" };
		user.set(mockUser);
		expect(get(user)).toEqual(mockUser);
	});

	it("should reset stores on logout", () => {
		isAuthenticated.set(true);
		user.set({ id: "1", email: "test@example.com", name: "Test User" });

		logout();

		expect(get(isAuthenticated)).toBe(false);
		expect(get(user)).toBeNull();
	});
});