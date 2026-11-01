<script lang="ts">
  import "../app.css";
  import "#lib/i18n.js";
  import { browser } from "$app/env";
  import { page } from "$app/state";
  import { onMount, type Snippet } from "svelte";
  import { refreshAll } from "$app/navigation";
  import Head from "#lib/components/atoms/Head.svelte";
  import NavigationBar from "#lib/components/molecules/NavigationBar.svelte";
  import QuickNavigation from "#lib/components/molecules/QuickNavigation.svelte";
  import Footer from "#lib/components/organisms/Footer.svelte";
  import { summaryVisibility } from "#lib/summary-visibility-store.js";
  import { autoPhraseEnabled } from "#lib/auto-phrase-store.js";
  import type { LayoutData } from "./$types";

  let { data, children }: { data: LayoutData; children: Snippet } = $props();

  let isAuthenticated = $derived(data.isAuthenticated);
  let isAuthPage = $derived(
    page.url.pathname === "/login" ||
      page.url.pathname === "/register" ||
      page.url.pathname === "/oauth/authorize",
  );

  let lastActiveTime = Date.now();
  // refreshAll() の重複実行を防ぐフラグ
  let isRefreshing = false;

  // visibilitychange のみでトークンチェックを行う。
  // focus イベントと併用すると「タブを戻した時」に両方が同時に発火して
  // refreshAll() が2回呼ばれ、並行リフレッシュ競合が起きるため focus は使わない。
  function handleVisibilityChange() {
    if (!browser || document.visibilityState !== "visible") return;
    const inactiveTime = Date.now() - lastActiveTime;
    lastActiveTime = Date.now();
    // 5分以上非アクティブかつ認証済みの場合のみリフレッシュ
    if (
      inactiveTime > 5 * 60 * 1000 &&
      isAuthenticated &&
      !isAuthPage &&
      !isRefreshing
    ) {
      isRefreshing = true;
      refreshAll().finally(() => {
        isRefreshing = false;
      });
    }
  }

  onMount(() => {
    // タイムゾーンオフセットをCookieに保存（分単位、負の値はUTCより進んでいる）
    const timezoneOffset = new Date().getTimezoneOffset();
    // biome-ignore lint/suspicious/noDocumentCookie: サーバーにタイムゾーン情報を送信するために必要
    document.cookie = `tz_offset=${timezoneOffset}; path=/; max-age=31536000`; // 1年間有効

    // ストアを初期化（アプリケーション全体で一度だけ）
    summaryVisibility.init();

    // auto-phraseの状態をbody要素のクラスに反映
    const unsubscribe = autoPhraseEnabled.subscribe((enabled) => {
      if (enabled) {
        document.body.classList.add("auto-phrase-enabled");
      } else {
        document.body.classList.remove("auto-phrase-enabled");
      }
    });

    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      unsubscribe();
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  });
</script>

<Head />

<div class="min-h-screen bg-gray-50 dark:bg-gray-800 transition-colors flex flex-col">
	<NavigationBar {isAuthenticated} {isAuthPage} />

	{#if isAuthenticated && !isAuthPage}
		<QuickNavigation />
	{/if}

	<main class="{isAuthenticated && !isAuthPage ? 'container mx-auto py-8' : ''} text-gray-900 dark:text-gray-100 flex-1">
		{@render children()}
	</main>

	<Footer {isAuthenticated} />
</div>

