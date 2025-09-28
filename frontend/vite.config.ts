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
			workbox: {
				globPatterns: ["**/*.{js,css,html,ico,png,svg}"],
				runtimeCaching: [
					{
						urlPattern: ({ url }) => {
							// 環境変数からAPI URLを取得（開発・本番環境共通）
							const apiBase = process.env.VITE_API_URL || "";
							return apiBase
								? url.href.startsWith(apiBase)
								: url.pathname.startsWith("/api");
						},
						handler: "NetworkFirst",
						options: {
							cacheName: "api-cache",
							cacheableResponse: { statuses: [0, 200] },
							networkTimeoutSeconds: 10,
						},
					},
				],
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
