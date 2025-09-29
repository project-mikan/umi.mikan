import adapter from "@sveltejs/adapter-auto";
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
		// adapter-auto only supports some environments, see https://svelte.dev/docs/kit/adapter-auto for a list.
		// If your environment is not supported, or you settled on a specific environment, switch out the adapter.
		// See https://svelte.dev/docs/kit/adapters for more information about adapters.
		adapter: adapter(),
		csp: {
			mode: "hash",
			directives: {
				"default-src": ["self"],
				"script-src": ["self"],
				"style-src": ["self", "unsafe-inline"],
				"img-src": ["self", "data:", "blob:"],
				"font-src": ["self"],
				"connect-src": ["self", "http://localhost:2001", "http://backend:8080"],
				"form-action": ["self"],
				"frame-ancestors": ["none"],
				"object-src": ["none"],
				"base-uri": ["self"]
			}
		}
	},
};

export default config;
