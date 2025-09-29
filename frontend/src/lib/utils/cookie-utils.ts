import { dev } from "$app/environment";

export interface CookieOptions {
	path: string;
	httpOnly?: boolean;
	secure?: boolean;
	sameSite?: "strict" | "lax" | "none";
	maxAge?: number;
}

export function getSecureCookieOptions(maxAge: number): CookieOptions {
	// 本番環境ではsecure=true、開発環境ではsecure=false
	return {
		path: "/",
		httpOnly: true,
		secure: !dev, // 本番環境では自動的にtrue
		sameSite: "strict",
		maxAge,
	};
}

// アクセストークン用のクッキー設定（15分）
export const ACCESS_TOKEN_COOKIE_OPTIONS = getSecureCookieOptions(60 * 15);

// リフレッシュトークン用のクッキー設定（30日）
export const REFRESH_TOKEN_COOKIE_OPTIONS = getSecureCookieOptions(
	60 * 60 * 24 * 30,
);
