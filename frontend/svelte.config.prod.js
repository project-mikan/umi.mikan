import adapter from "@sveltejs/adapter-node";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		alias: {
			"$src/*": "src/*",
			"$apiSchema/*": "src/apiSchema/*",
		},
		adapter: adapter(),
		files: {
			hooks: {
				server: "src/lib/hooks/hooks.server.ts",
			},
		},
	},
};

export default config;
