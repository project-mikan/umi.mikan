import adapter from "@sveltejs/adapter-node";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vitest/config";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [
    tailwindcss(),
    // SvelteKit 3 から svelte.config.js は使われず、設定はここで sveltekit() に渡す。
    // 本番は adapter-node で動かす（vite dev はアダプターを使わないので開発環境にも影響しない）
    sveltekit({
      preprocess: vitePreprocess(),
      adapter: adapter(),
      csp: {
        mode: "hash",
        directives: {
          "default-src": ["self"],
          "script-src": ["self"],
          "style-src": ["self", "unsafe-inline"],
          "img-src": ["self", "data:", "blob:"],
          "font-src": ["self"],
          // backend:8080 は Docker 内部のホスト名でブラウザから到達できないため指定しない
          // （SvelteKit 3 の型はドットを含まないホスト名を受け付けない）
          "connect-src": ["self", "http://localhost:2001"],
          "form-action": ["self"],
          "frame-ancestors": ["none"],
          "object-src": ["none"],
          "base-uri": ["self"],
        },
      },
    }),
  ],
  server: {
    // DockerのボリュームマウントでHMRが動作するようにpollingを使用。
    // ポーリングは監視ファイル全件をstatし続けるため、巨大なキャッシュディレクトリが
    // 監視対象に入るとlibuvスレッドプールが飽和してSSRが数秒〜無応答になる。
    // (.pnpm-storeを監視していたことが原因の障害が過去にあったため明示的に除外)
    watch: {
      usePolling: true,
      interval: 300,
      ignored: ["**/.pnpm-store/**", "**/dev-dist/**"],
    },
  },
  ssr: {
    noExternal: [],
    external: ["chart.js"],
  },
  // Svelteコンポーネントをvitestでマウントする際、クライアント向けmount()を使わせるために必要
  // （未設定だとvitestがSSRモードで依存解決しSvelteのサーバー用mount()が使われてしまい、
  //   "mount(...) is not available on the server" エラーになる）
  resolve: {
    conditions: ["browser"],
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
