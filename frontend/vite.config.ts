import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vitest/config";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [tailwindcss(), sveltekit()],
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
