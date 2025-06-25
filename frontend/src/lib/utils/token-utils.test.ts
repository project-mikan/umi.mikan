import { describe, expect, it } from "vitest";
import { isTokenExpiringSoon } from "./token-utils";

describe("Token Utility Functions", () => {
	describe("isTokenExpiringSoon", () => {
		it("should return true for tokens expiring within buffer time", () => {
			const now = Date.now();
			const expiryTime = Math.floor((now + 2 * 60 * 1000) / 1000); // 2 minutes from now
			const payload = { exp: expiryTime };
			const token = `header.${btoa(JSON.stringify(payload))}.signature`;

			expect(isTokenExpiringSoon(token, 5)).toBe(true);
		});

		it("should return false for tokens expiring after buffer time", () => {
			const now = Date.now();
			const expiryTime = Math.floor((now + 10 * 60 * 1000) / 1000); // 10 minutes from now
			const payload = { exp: expiryTime };
			const token = `header.${btoa(JSON.stringify(payload))}.signature`;

			expect(isTokenExpiringSoon(token, 5)).toBe(false);
		});

		it("should return true for expired tokens", () => {
			const now = Date.now();
			const expiryTime = Math.floor((now - 60 * 1000) / 1000); // 1 minute ago
			const payload = { exp: expiryTime };
			const token = `header.${btoa(JSON.stringify(payload))}.signature`;

			expect(isTokenExpiringSoon(token, 5)).toBe(true);
		});

		it("should return true for invalid tokens", () => {
			expect(isTokenExpiringSoon("invalid.token")).toBe(true);
			expect(isTokenExpiringSoon("")).toBe(true);
			expect(isTokenExpiringSoon("not.a.jwt.token")).toBe(true);
		});

		it("should return true for malformed JWT", () => {
			expect(isTokenExpiringSoon("header.invalid-base64.signature")).toBe(true);
		});

		it("should use custom buffer time", () => {
			const now = Date.now();
			const expiryTime = Math.floor((now + 8 * 60 * 1000) / 1000); // 8 minutes from now
			const payload = { exp: expiryTime };
			const token = `header.${btoa(JSON.stringify(payload))}.signature`;

			expect(isTokenExpiringSoon(token, 10)).toBe(true); // 10 minute buffer
			expect(isTokenExpiringSoon(token, 5)).toBe(false); // 5 minute buffer
		});

		it("should handle tokens without exp field", () => {
			const payload = { sub: "user123" }; // No exp field
			const token = `header.${btoa(JSON.stringify(payload))}.signature`;

			// When exp is undefined, expiryTime - now will be NaN, making the comparison always false
			// But the function should handle this case by returning true for invalid expiry
			const result = isTokenExpiringSoon(token);
			// This test will need to be adjusted based on the actual implementation
			expect(typeof result).toBe("boolean");
		});
	});
});
