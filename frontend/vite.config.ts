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
			},
			injectRegister: "auto",
			mode:
				process.env.NODE_ENV === "development" ? "development" : "production",
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
					{
						// SvelteKitのナビゲーションリクエスト(HTMLページ)をキャッシュ
						urlPattern: ({ request }) => request.mode === "navigate",
						handler: "NetworkFirst",
						options: {
							cacheName: "pages-cache",
							cacheableResponse: { statuses: [0, 200] },
							networkTimeoutSeconds: 3,
						},
					},
				],
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
