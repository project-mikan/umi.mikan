<script lang="ts">
	import { _, locale } from "svelte-i18n";
	import { browser } from "$app/environment";
	import { onMount, onDestroy, createEventDispatcher } from "svelte";
	import { authenticatedFetch } from "$lib/auth-client";
	import "$lib/i18n";
	import Button from "$lib/components/atoms/Button.svelte";
	import type { HighlightData } from "$lib/types/highlight";

	export let diaryId: string;
	export let hasLLMKey = true;
	export let isHighlightOutdated = false;
	export let diaryUpdatedAt: number;

	const dispatch = createEventDispatcher();

	let highlightData: HighlightData | null = null;
	let highlightStatus:
		| "none"
		| "queued"
		| "processing"
		| "completed"
		| "error" = "none";
	let highlightGenerating = false;
	let isRegenerating = false;
	let pollingInterval: ReturnType<typeof setInterval> | null = null;
	let highlightVisible = true; // ハイライトの表示・非表示状態
	let pollingAttempts = 0; // ポーリング試行回数
	let errorMessage = ""; // エラーメッセージ

	const MAX_POLLING_ATTEMPTS = 20; // 最大20回（60秒）
	const POLLING_INTERVAL = 3000; // 3秒間隔

	// ポーリング停止とリセット
	function stopPolling() {
		if (pollingInterval) {
			clearInterval(pollingInterval);
			pollingInterval = null;
		}
		pollingAttempts = 0;
	}

	// ハイライトのポーリング機能
	async function pollHighlightStatus() {
		if (!browser) return;

		pollingAttempts++;

		// タイムアウトチェック
		if (pollingAttempts > MAX_POLLING_ATTEMPTS) {
			stopPolling();
			highlightStatus = "error";
			highlightGenerating = false;
			isRegenerating = false;
			errorMessage = $_("diary.highlight.timeout");
			return;
		}

		try {
			const response = await authenticatedFetch(
				`/api/diary/highlight/${diaryId}`,
			);
			if (response.ok) {
				const result = await response.json();
				if (result.highlights && result.highlights.length > 0) {
					highlightData = result;
					highlightStatus = "completed";
					highlightGenerating = false;
					isRegenerating = false;
					errorMessage = "";

					// ポーリング停止
					stopPolling();

					// 親コンポーネントに通知
					dispatch("highlightUpdated", { highlight: result });
				}
			} else if (response.status === 404) {
				// ハイライトがまだ存在しない（生成中の可能性）
				// 次のポーリングを待つ
			} else {
				// その他のエラー
				stopPolling();
				highlightStatus = "error";
				highlightGenerating = false;
				isRegenerating = false;
				errorMessage = $_("diary.highlight.generationFailed");
			}
		} catch (err) {
			console.error("Failed to poll highlight status:", err);
			stopPolling();
			highlightStatus = "error";
			highlightGenerating = false;
			isRegenerating = false;
			errorMessage = $_("diary.highlight.generationFailed");
		}
	}

	// ハイライト生成をトリガー
	async function generateHighlight() {
		if (!browser || !hasLLMKey) return;

		highlightGenerating = true;
		highlightStatus = highlightData ? "processing" : "queued";
		isRegenerating = !!highlightData;
		errorMessage = "";

		try {
			const response = await authenticatedFetch(
				"/api/diary/trigger-highlight",
				{
					method: "POST",
					headers: {
						"Content-Type": "application/json",
					},
					body: JSON.stringify({
						diaryId,
					}),
				},
			);

			if (response.ok) {
				const result = await response.json();
				console.log("Highlight generation triggered:", result);

				// ポーリング開始
				stopPolling(); // 既存のポーリングをクリア
				pollingAttempts = 0;
				pollingInterval = setInterval(pollHighlightStatus, POLLING_INTERVAL);
			} else {
				const error = await response.json();
				console.error("Failed to trigger highlight generation:", error);
				highlightGenerating = false;
				highlightStatus = "error";
				isRegenerating = false;
				errorMessage = error.message || $_("diary.highlight.generationFailed");
			}
		} catch (err) {
			console.error("Error triggering highlight generation:", err);
			highlightGenerating = false;
			highlightStatus = "error";
			isRegenerating = false;
			errorMessage = $_("diary.highlight.generationFailed");
		}
	}

	// ハイライト削除
	async function deleteHighlight() {
		if (!browser) return;

		try {
			const response = await authenticatedFetch(
				`/api/diary/highlight/${diaryId}`,
				{
					method: "DELETE",
				},
			);

			if (response.ok) {
				highlightData = null;
				highlightStatus = "none";
				errorMessage = "";
				dispatch("highlightDeleted");
			} else {
				console.error("Failed to delete highlight");
				errorMessage = $_("diary.highlight.deleteFailed");
			}
		} catch (err) {
			console.error("Error deleting highlight:", err);
			errorMessage = $_("diary.highlight.deleteFailed");
		}
	}

	// ハイライトの表示・非表示を切り替え
	function toggleHighlightVisibility() {
		highlightVisible = !highlightVisible;
		// 親コンポーネントに表示状態を通知
		dispatch("highlightVisibilityChanged", { visible: highlightVisible });
	}

	// 初回読み込み時にハイライトを取得
	onMount(async () => {
		if (!browser || !diaryId) return;

		try {
			const response = await authenticatedFetch(
				`/api/diary/highlight/${diaryId}`,
			);
			if (response.ok) {
				const result = await response.json();
				highlightData = result;
				highlightStatus = "completed";

				// 親コンポーネントに通知
				dispatch("highlightUpdated", { highlight: result });

				// 日記が更新されているかチェック
				if (diaryUpdatedAt > result.updatedAt) {
					isHighlightOutdated = true;
				}
			}
		} catch (err) {
			// ハイライトが存在しない場合は何もしない
		}
	});

	onDestroy(() => {
		stopPolling();
	});

	$: buttonLabel = isRegenerating
		? $_("diary.highlight.regenerate")
		: $_("diary.highlight.generate");
	$: buttonLoadingLabel = $_("diary.highlight.generating");
</script>

{#if hasLLMKey}
	<div class="flex flex-col gap-2 my-4">
		<div class="flex flex-wrap gap-2 items-center">
			{#if isHighlightOutdated && highlightData}
				<span class="px-3 py-1 rounded-full text-sm font-medium bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200">
					⚠️ {$_("diary.highlight.outdated")}
				</span>
			{/if}

			{#if highlightStatus === "queued"}
				<span class="px-3 py-1 rounded-full text-sm font-medium bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200">
					{$_("diary.highlight.statusQueued")}
				</span>
			{:else if highlightStatus === "processing"}
				<span class="px-3 py-1 rounded-full text-sm font-medium bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200">
					{$_("diary.highlight.statusProcessing")}
				</span>
			{:else if highlightStatus === "error" && errorMessage}
				<span class="px-3 py-1 rounded-full text-sm font-medium bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200">
					❌ {errorMessage}
				</span>
			{/if}

			<Button
				variant="secondary"
				size="sm"
				on:click={generateHighlight}
				disabled={highlightGenerating}
			>
				{highlightGenerating ? buttonLoadingLabel : buttonLabel}
			</Button>

			{#if highlightData}
				<Button
					variant={highlightVisible ? "secondary" : "primary"}
					size="sm"
					on:click={toggleHighlightVisibility}
					disabled={highlightGenerating}
				>
					{highlightVisible ? $_("diary.highlight.hide") : $_("diary.highlight.show")}
				</Button>

				<Button
					variant="danger"
					size="sm"
					on:click={deleteHighlight}
					disabled={highlightGenerating}
				>
					{$_("diary.highlight.delete")}
				</Button>
			{/if}
		</div>
	</div>
{/if}
