import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	define: {
		global: 'globalThis',
	},
	ssr: {
		noExternal: ['@grpc/grpc-js', '@grpc/proto-loader'],
		external: ['fs', 'path', 'url']
	},
	optimizeDeps: {
		exclude: ['@grpc/grpc-js', '@grpc/proto-loader']
	}
});
