<script lang="ts">
import { onMount } from "svelte";
import { _ } from "svelte-i18n";
import "$lib/i18n";
import type { BeforeInstallPromptEvent } from "../pwa-types";

let deferredPrompt: BeforeInstallPromptEvent | null = null;
let canInstall = false;
let isInstalled = false;

onMount(() => {
	// Check if already installed
	isInstalled = window.matchMedia("(display-mode: standalone)").matches;

	if (isInstalled) {
		return;
	}

	// Note: Always show install button regardless of previous dismissals

	// Listen for beforeinstallprompt event
	window.addEventListener("beforeinstallprompt", (e) => {
		console.log("beforeinstallprompt event fired");
		e.preventDefault();
		deferredPrompt = e;
		canInstall = true;
	});

	// Check if PWA is already installable via other means
	if ("serviceWorker" in navigator && "BeforeInstallPromptEvent" in window) {
		canInstall = true;
	}

	// For browsers that support PWA but don't fire beforeinstallprompt immediately
	setTimeout(() => {
		if (!canInstall && !isInstalled) {
			console.log("Enabling install button as fallback");
			canInstall = true;
		}
	}, 2000);
});

const handleInstall = async () => {
	if (deferredPrompt) {
		// Use native browser install prompt
		console.log("Using native install prompt");
		deferredPrompt.prompt();
		const { outcome } = await deferredPrompt.userChoice;

		if (outcome === "accepted") {
			console.log("PWA installation accepted");
			localStorage.setItem("pwa-installed", "true");
			canInstall = false;
		} else {
			console.log("PWA installation dismissed");
		}

		deferredPrompt = null;
		return;
	}

	// Fallback: Show manual installation instructions
	console.log("Showing manual installation instructions");

	// Detect browser type for specific instructions
	const userAgent = navigator.userAgent.toLowerCase();
	let instructions = $_("pwa.install.manualInstructions");

	if (userAgent.includes("chrome") && !userAgent.includes("edg")) {
		instructions =
			"Chrome: 右上の⋮メニュー → 「アプリをインストール」を選択してください。";
	} else if (userAgent.includes("edg")) {
		instructions =
			"Edge: 右上の⋯メニュー → 「アプリ」→ 「このサイトをアプリとしてインストール」を選択してください。";
	} else if (userAgent.includes("firefox")) {
		instructions =
			"Firefox: アドレスバーの🏠アイコンをクリックするか、右上のメニューから「ホーム画面に追加」を選択してください。";
	} else if (userAgent.includes("safari")) {
		instructions =
			"Safari: 下部の共有ボタン📤 → 「ホーム画面に追加」を選択してください。";
	}

	alert(instructions);
};

const handleDismiss = () => {
	// Do nothing - keep the button visible
	// User can always access install button from home page
};
</script>

{#if canInstall && !isInstalled}
	<div class="mt-8 flex justify-center">
		<button
			on:click={handleInstall}
			class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-6 rounded-lg transition-colors flex items-center"
		>
			<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
			</svg>
			{$_("pwa.install.install")}
		</button>
	</div>
{/if}