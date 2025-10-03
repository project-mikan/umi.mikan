<script lang="ts">
import "../app.css";
import "$lib/i18n";
import { browser } from "$app/environment";
import { page } from "$app/stores";
import { onMount, onDestroy } from "svelte";
import { invalidateAll } from "$app/navigation";
import Head from "$lib/components/atoms/Head.svelte";
import NavigationBar from "$lib/components/molecules/NavigationBar.svelte";
import QuickNavigation from "$lib/components/molecules/QuickNavigation.svelte";
import Footer from "$lib/components/organisms/Footer.svelte";
import { summaryVisibility } from "$lib/summary-visibility-store";
import { autoPhraseEnabled } from "$lib/auto-phrase-store";
import PWAInstallPrompt from "$lib/components/PWAInstallPrompt.svelte";
import PWAUpdateNotification from "$lib/components/PWAUpdateNotification.svelte";
import type { LayoutData } from "./$types";

export let data: LayoutData;

$: isAuthenticated = data.isAuthenticated;
$: isAuthPage =
	$page.url.pathname === "/login" || $page.url.pathname === "/register";

let lastActiveTime = Date.now();

// ページがフォーカスされた時にトークンをチェック
function handleVisibilityChange() {
	if (browser && document.visibilityState === "visible") {
		const inactiveTime = Date.now() - lastActiveTime;
		// 5分以上非アクティブだった場合、トークンをリフレッシュ
		if (inactiveTime > 5 * 60 * 1000 && isAuthenticated && !isAuthPage) {
			invalidateAll();
		}
		lastActiveTime = Date.now();
	}
}

// ページがフォーカスされた時
function handleFocus() {
	if (!browser) return;
	const inactiveTime = Date.now() - lastActiveTime;
	// 5分以上非アクティブだった場合、トークンをリフレッシュ
	if (inactiveTime > 5 * 60 * 1000 && isAuthenticated && !isAuthPage) {
		invalidateAll();
	}
	lastActiveTime = Date.now();
}

onMount(() => {
	// ストアを初期化（アプリケーション全体で一度だけ）
	summaryVisibility.init();

	// auto-phraseの状態をbody要素のクラスに反映
	const unsubscribe = autoPhraseEnabled.subscribe((enabled) => {
		if (browser) {
			if (enabled) {
				document.body.classList.add("auto-phrase-enabled");
			} else {
				document.body.classList.remove("auto-phrase-enabled");
			}
		}
	});

	// visibilitychangeイベントをリッスン
	document.addEventListener("visibilitychange", handleVisibilityChange);
	window.addEventListener("focus", handleFocus);

	return () => {
		unsubscribe();
	};
});

onDestroy(() => {
	// ブラウザ環境でのみイベントリスナーを削除
	if (browser) {
		document.removeEventListener("visibilitychange", handleVisibilityChange);
		window.removeEventListener("focus", handleFocus);
	}
});
</script>

<Head />

<div class="min-h-screen bg-gray-50 dark:bg-gray-800 transition-colors flex flex-col">
	<NavigationBar {isAuthenticated} {isAuthPage} />

	{#if isAuthenticated && !isAuthPage}
		<QuickNavigation />
	{/if}

	<main class="{isAuthenticated && !isAuthPage ? 'container mx-auto py-8' : ''} text-gray-900 dark:text-gray-100 flex-1">
		<slot />
	</main>

	<Footer {isAuthenticated} />
</div>

<!-- PWA Components -->
<PWAInstallPrompt />
<PWAUpdateNotification />