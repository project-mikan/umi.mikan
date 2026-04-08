<script lang="ts">
  import { _ } from "svelte-i18n";
  import "$lib/i18n";

  export let embeddingStatus: Promise<{
    indexed: boolean;
    modelVersion: string;
    createdAt: number;
    updatedAt: number;
    chunkCount: number;
    chunkSummaries: string[];
  } | null> | null;

  // モバイル用トグル（デフォルトは非表示）
  let mobileOpen = false;
</script>

{#await embeddingStatus}
  <!-- ローディング中は何も表示しない -->
{:then status}
  {#if status?.indexed && status.chunkSummaries.length > 0}
    <div class="rounded-xl border border-indigo-200 dark:border-indigo-800/50 bg-white dark:bg-gray-800/50 shadow-sm overflow-hidden">
      <!-- ヘッダー行 -->
      <div class="flex items-center justify-between px-4 py-2.5 bg-indigo-50 dark:bg-indigo-950/30 border-b border-indigo-100 dark:border-indigo-800/40">
        <span class="text-xs font-semibold text-indigo-700 dark:text-indigo-300 tracking-wide uppercase">
          {$_("diary.chunkTimeline.title")}
        </span>
        <!-- モバイルのみトグルボタン表示 -->
        <button
          type="button"
          on:click={() => (mobileOpen = !mobileOpen)}
          class="sm:hidden flex items-center gap-1 text-xs text-indigo-500 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 transition-colors"
          aria-expanded={mobileOpen}
        >
          {mobileOpen ? $_("diary.chunkTimeline.hide") : $_("diary.chunkTimeline.show")}
          <svg
            class="w-3.5 h-3.5 transition-transform {mobileOpen ? 'rotate-180' : ''}"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>

      <!-- タイムライン本体: PC は常時表示、モバイルはトグル -->
      <div class="{mobileOpen ? '' : 'hidden'} sm:block px-4 pt-3 pb-1">
        <!-- チャンク一覧 -->
        <div class="border-l-2 border-indigo-200 dark:border-indigo-700 pl-5 space-y-2 pb-2">
          {#each status.chunkSummaries as summary}
            <div class="relative">
              <div class="absolute -left-[1.6rem] top-1 w-2.5 h-2.5 rounded-full border-2 border-indigo-400 dark:border-indigo-500 bg-white dark:bg-gray-800"></div>
              <p class="text-sm text-gray-700 dark:text-gray-300 leading-snug">
                {summary || $_("diary.embedding.chunkSummaryEmpty")}
              </p>
            </div>
          {/each}
        </div>
        <!-- メタ情報フッター -->
        <div class="mt-1 pb-2 flex flex-wrap gap-x-4 gap-y-0.5 text-[11px] text-gray-400 dark:text-gray-500">
          <span>{$_("diary.embedding.modelVersion")}: <span class="font-mono">{status.modelVersion}</span></span>
          <span>{$_("diary.embedding.updatedAt")}: {new Date(status.updatedAt * 1000).toLocaleString()}</span>
        </div>
      </div>
    </div>
  {/if}
{/await}
