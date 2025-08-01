import { get } from "svelte/store";
import { beforeEach, describe, expect, it } from "vitest";
import { isAuthenticated, logout, type User, user } from "./auth";

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
		const mockUser: User = {
			id: "1",
			email: "test@example.com",
			name: "Test User",
		};
		user.set(mockUser);
		expect(get(user)).toEqual(mockUser);
	});

	it("should reset stores on logout", () => {
		isAuthenticated.set(true);
		const mockUser: User = {
			id: "1",
			email: "test@example.com",
			name: "Test User",
		};
		user.set(mockUser);

		logout();

		expect(get(isAuthenticated)).toBe(false);
		expect(get(user)).toBeNull();
	});
});
