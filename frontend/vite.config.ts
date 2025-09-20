import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	ssr: {
		noExternal: [],
		external: ["chart.js"]
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
