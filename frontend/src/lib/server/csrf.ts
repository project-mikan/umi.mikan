import { randomBytes } from "node:crypto";
import type { Cookies } from "@sveltejs/kit";

const CSRF_TOKEN_LENGTH = 32;
const CSRF_COOKIE_NAME = "csrfToken";

/**
 * CSRFトークンを生成します
 */
export function generateCSRFToken(): string {
	return randomBytes(CSRF_TOKEN_LENGTH).toString("hex");
}

/**
 * CSRFトークンをクッキーに設定します
 */
export function setCSRFToken(cookies: Cookies): string {
	const token = generateCSRFToken();
	cookies.set(CSRF_COOKIE_NAME, token, {
		path: "/",
		httpOnly: false, // JSからアクセス可能にする必要がある
		secure: process.env.NODE_ENV === "production",
		sameSite: "strict",
		maxAge: 60 * 60 * 24, // 24時間
	});
	return token;
}

/**
 * CSRFトークンを取得します
 */
export function getCSRFToken(cookies: Cookies): string | undefined {
	return cookies.get(CSRF_COOKIE_NAME);
}

/**
 * CSRFトークンを検証します
 */
export function validateCSRFToken(
	cookies: Cookies,
	submittedToken: string | null,
): boolean {
	if (!submittedToken) {
		return false;
	}

	const storedToken = getCSRFToken(cookies);
	if (!storedToken) {
		return false;
	}

	// タイミング攻撃を防ぐため、固定時間で比較
	return timingSafeEqual(storedToken, submittedToken);
}

/**
 * タイミング攻撃を防ぐ文字列比較
 */
function timingSafeEqual(a: string, b: string): boolean {
	if (a.length !== b.length) {
		return false;
	}

	let result = 0;
	for (let i = 0; i < a.length; i++) {
		result |= a.charCodeAt(i) ^ b.charCodeAt(i);
	}

	return result === 0;
}
