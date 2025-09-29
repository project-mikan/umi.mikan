import type { Handle } from "@sveltejs/kit";

export const handle: Handle = async ({ event, resolve }) => {
	const response = await resolve(event);

	// Content Security Policy の設定
	const cspDirectives = [
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline'", // SvelteKitでは'unsafe-inline'が必要な場合がある
		"style-src 'self' 'unsafe-inline'", // インラインスタイル用
		"img-src 'self' data: blob:",
		"font-src 'self'",
		"connect-src 'self' http://localhost:2001 http://backend:8080", // gRPCバックエンドへの接続
		"form-action 'self'",
		"frame-ancestors 'none'",
		"object-src 'none'",
		"base-uri 'self'",
		"upgrade-insecure-requests"
	];

	// 本番環境ではより厳しいCSPを設定
	if (process.env.NODE_ENV === "production") {
		cspDirectives[5] = "connect-src 'self' https://backend:8080"; // HTTPSのみ
	}

	response.headers.set("Content-Security-Policy", cspDirectives.join("; "));

	// その他のセキュリティヘッダー
	response.headers.set("X-Frame-Options", "DENY");
	response.headers.set("X-Content-Type-Options", "nosniff");
	response.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");
	response.headers.set("Permissions-Policy", "camera=(), microphone=(), geolocation=()");

	return response;
};