import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";

export default defineConfig({
	plugins: [sveltekit()],
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
