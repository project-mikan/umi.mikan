import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import { SvelteKitPWA } from "@vite-pwa/sveltekit";

// Note: Manifest is now generated dynamically via API route at /manifest.webmanifest

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		SvelteKitPWA({
			registerType: "autoUpdate",
			devOptions: {
				enabled: true,
				type: "module",
				navigateFallback: "/",
			},
			injectRegister: "auto",
			mode: process.env.NODE_ENV === "development" ? "development" : "production",
			workbox: {
				globPatterns: ["**/*.{js,css,html,ico,png,svg}"],
				runtimeCaching: [
					{
						urlPattern: ({ url }) => {
							// SvelteKit API routesをキャッシュ対象とする
							return url.pathname.startsWith("/api/");
						},
						handler: "NetworkFirst",
						options: {
							cacheName: "api-cache",
							cacheableResponse: { statuses: [0, 200] },
							networkTimeoutSeconds: 10,
						},
					},
				],
				// 本番環境での静的ファイル配信を改善
				navigateFallback: "/",
				navigateFallbackAllowlist: [/^(?!\/__).*/], // __で始まるパス以外
				// 本番環境での動的URLマッチングを無効にし、相対パスを使用
				cleanupOutdatedCaches: true,
			},
			manifest: false, // Disable static manifest generation - use dynamic API route instead
		}),
	],
	ssr: {
		noExternal: [],
		external: ["chart.js"],
	},
	test: {
		include: ["src/**/*.{test,spec}.{js,ts}"],
		environment: "jsdom",
		setupFiles: ["./src/test/setup.ts"],
		server: {
			deps: {
				inline: ["@testing-library/svelte"],
			},
		},
	},
});
