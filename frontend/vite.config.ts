import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";
import { SvelteKitPWA } from "@vite-pwa/sveltekit";

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		SvelteKitPWA({
			registerType: "autoUpdate",
			workbox: {
				globPatterns: ["**/*.{js,css,html,ico,png,svg}"],
				runtimeCaching: [
					{
						urlPattern: /^http:\/\/localhost:2001\/.*$/,
						handler: "NetworkFirst",
						options: {
							cacheName: "api-cache",
							cacheableResponse: { statuses: [0, 200] },
							networkTimeoutSeconds: 10,
						},
					},
				],
			},
			manifest: {
				name: "umi.mikan - 日記アプリ",
				short_name: "umi.mikan",
				description: "毎日使う日記アプリ",
				start_url: "/",
				display: "standalone",
				background_color: "#ffffff",
				theme_color: "#3b82f6",
				orientation: "portrait-primary",
				categories: ["lifestyle", "productivity"],
				lang: "ja",
				icons: [
					{
						src: "icons/icon-72x72.png",
						sizes: "72x72",
						type: "image/png",
					},
					{
						src: "icons/icon-96x96.png",
						sizes: "96x96",
						type: "image/png",
					},
					{
						src: "icons/icon-128x128.png",
						sizes: "128x128",
						type: "image/png",
					},
					{
						src: "icons/icon-144x144.png",
						sizes: "144x144",
						type: "image/png",
					},
					{
						src: "icons/icon-152x152.png",
						sizes: "152x152",
						type: "image/png",
					},
					{
						src: "icons/icon-192x192.png",
						sizes: "192x192",
						type: "image/png",
					},
					{
						src: "icons/icon-384x384.png",
						sizes: "384x384",
						type: "image/png",
					},
					{
						src: "icons/icon-512x512.png",
						sizes: "512x512",
						type: "image/png",
					},
					{
						src: "icons/icon-192x192.png",
						sizes: "192x192",
						type: "image/png",
						purpose: "maskable",
					},
					{
						src: "icons/icon-512x512.png",
						sizes: "512x512",
						type: "image/png",
						purpose: "maskable",
					},
				],
			},
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
