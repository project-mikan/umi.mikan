<script lang="ts">
	import { _, locale } from "svelte-i18n";
	import { browser } from "$app/environment";
	import { onMount, onDestroy, createEventDispatcher } from "svelte";
	import { authenticatedFetch } from "$lib/auth-client";
	import "$lib/i18n";
	import Button from "$lib/components/atoms/Button.svelte";

	interface Highlight {
		start: number;
		end: number;
		text: string;
	}

	interface HighlightData {
		highlights: Highlight[];
		createdAt: number;
		updatedAt: number;
	}

	export let diaryId: string;
	export let hasLLMKey = true;
	export let isHighlightOutdated = false;
	export let diaryUpdatedAt: number;

	const dispatch = createEventDispatcher();

	let highlightData: HighlightData | null = null;
	let highlightStatus: "none" | "queued" | "processing" | "completed" = "none";
	let highlightGenerating = false;
	let isRegenerating = false;
	let pollingInterval: ReturnType<typeof setInterval> | null = null;

	// ハイライトのポーリング機能
	async function pollHighlightStatus() {
		if (!browser) return;
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

					// ポーリング停止
					if (pollingInterval) {
						clearInterval(pollingInterval);
						pollingInterval = null;
					}

					// 親コンポーネントに通知
					dispatch("highlightUpdated", { highlight: result });
				}
			} else if (response.status === 404) {
				// ハイライトがまだ存在しない（生成中の可能性）
			}
		} catch (err) {
			console.error("Failed to poll highlight status:", err);
		}
	}

	// ハイライト生成をトリガー
	async function generateHighlight() {
		if (!browser || !hasLLMKey) return;

		highlightGenerating = true;
		highlightStatus = highlightData ? "processing" : "queued";
		isRegenerating = !!highlightData;

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
				if (pollingInterval) {
					clearInterval(pollingInterval);
				}
				pollingInterval = setInterval(pollHighlightStatus, 3000);
			} else {
				const error = await response.json();
				console.error("Failed to trigger highlight generation:", error);
				highlightGenerating = false;
				highlightStatus = highlightData ? "completed" : "none";
				isRegenerating = false;
			}
		} catch (err) {
			console.error("Error triggering highlight generation:", err);
			highlightGenerating = false;
			highlightStatus = highlightData ? "completed" : "none";
			isRegenerating = false;
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
				dispatch("highlightDeleted");
			} else {
				console.error("Failed to delete highlight");
			}
		} catch (err) {
			console.error("Error deleting highlight:", err);
		}
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
		if (pollingInterval) {
			clearInterval(pollingInterval);
		}
	});

	$: buttonLabel = isRegenerating
		? $_("diary.highlight.regenerate")
		: $_("diary.highlight.generate");
	$: buttonLoadingLabel = $_("diary.highlight.generating");
</script>

{#if hasLLMKey}
	<div class="flex flex-wrap gap-2 items-center my-4">
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
				variant="danger"
				size="sm"
				on:click={deleteHighlight}
				disabled={highlightGenerating}
			>
				{$_("diary.highlight.delete")}
			</Button>
		{/if}
	</div>
{/if}
