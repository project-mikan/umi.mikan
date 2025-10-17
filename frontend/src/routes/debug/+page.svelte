<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import { authenticatedFetch } from "$lib/auth-client";
import Head from "$lib/components/atoms/Head.svelte";
import Card from "$lib/components/atoms/Card.svelte";

let isTriggering = false;
let message = "";
let messageType: "success" | "error" | "" = "";

// トレンド分析を手動トリガー
async function triggerLatestTrend() {
	isTriggering = true;
	message = "";
	messageType = "";

	try {
		const response = await authenticatedFetch("/api/diary/trigger-latest-trend", {
			method: "POST",
		});

		if (response.ok) {
			const result = await response.json();
			message = result.message || "トレンド分析の生成をキューに追加しました";
			messageType = "success";
		} else {
			const errorData = await response.json().catch(() => ({}));
			message = errorData.message || "トレンド分析の生成に失敗しました";
			messageType = "error";
		}
	} catch (error) {
		console.error("Failed to trigger latest trend:", error);
		message = "トレンド分析の生成に失敗しました";
		messageType = "error";
	} finally {
		isTriggering = false;
	}
}
</script>

<Head title="Debug - umi.mikan" />

<div class="max-w-4xl mx-auto p-4 space-y-6">
	<!-- ページタイトル -->
	<div class="text-center">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
			{$_("navigation.debug")}
		</h1>
		<p class="text-red-600 dark:text-red-400 text-sm font-medium">
			開発環境専用ページ
		</p>
	</div>

	<!-- 直近トレンド分析トリガー -->
	<Card>
		<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
			直近トレンド分析 手動トリガー
		</h3>
		<p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
			過去7日間の日記から直近トレンド分析を即座に生成します。
		</p>

		<button
			type="button"
			on:click={triggerLatestTrend}
			disabled={isTriggering}
			class="bg-red-600 hover:bg-red-700 disabled:bg-red-300 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2"
		>
			{isTriggering ? "生成中..." : "トレンド分析を生成"}
		</button>

		{#if message}
			<div
				class="mt-4 p-4 rounded {messageType === 'success'
					? 'bg-green-100 border border-green-400 text-green-700'
					: 'bg-red-100 border border-red-400 text-red-700'}"
			>
				{message}
			</div>
		{/if}
	</Card>

	<!-- 説明 -->
	<Card>
		<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
			このページについて
		</h3>
		<p class="text-sm text-gray-600 dark:text-gray-400">
			このページは開発環境でのみ表示されます。本番環境では表示されません。
		</p>
	</Card>
</div>
