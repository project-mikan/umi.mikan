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
	export let content: string; // 日記の内容
	export let hasLLMKey = true;
	export let isHighlightOutdated = false; // ハイライトが古いかどうか
	export let diaryUpdatedAt: number; // 日記の更新日時

	const dispatch = createEventDispatcher();

	let highlightData: HighlightData | null = null;
	let highlightStatus: "none" | "queued" | "processing" | "completed" = "none";
	let highlightGenerating = false;
	let isRegenerating = false;
	let pollingInterval: ReturnType<typeof setInterval> | null = null;
	let highlightJustUpdated = false;
	let showHighlight = false;

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
					// ハイライトが完成した
					const oldHighlight = highlightData;
					highlightData = result;
					highlightStatus = "completed";
					highlightGenerating = false;
					isRegenerating = false;

					// ポーリング停止
					if (pollingInterval) {
						clearInterval(pollingInterval);
						pollingInterval = null;
					}

					// 更新アニメーション
					if (oldHighlight && oldHighlight.updatedAt !== result.updatedAt) {
						highlightJustUpdated = true;
						setTimeout(() => {
							highlightJustUpdated = false;
						}, 2000);
					}

					// 親コンポーネントに通知
					dispatch("highlightUpdated", { highlight: result });
				}
			} else if (response.status === 404) {
				// ハイライトがまだ存在しない（生成中の可能性）
				// ポーリング継続
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

				// 日記が更新されているかチェック
				if (diaryUpdatedAt > result.createdAt) {
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

	// ハイライトを適用したHTMLを生成
	function getHighlightedContent(): string {
		if (!highlightData || !highlightData.highlights || !content) {
			return content;
		}

		// ハイライト範囲をソート（開始位置の降順）
		const sortedHighlights = [...highlightData.highlights].sort(
			(a, b) => b.start - a.start,
		);

		let result = content;
		for (const highlight of sortedHighlights) {
			const before = result.substring(0, highlight.start);
			const highlighted = result.substring(highlight.start, highlight.end);
			const after = result.substring(highlight.end);
			result = `${before}<mark class="bg-yellow-300 dark:bg-yellow-600 px-1 rounded font-medium">${highlighted}</mark>${after}`;
		}

		return result;
	}

	$: buttonLabel = isRegenerating
		? $_("diary.highlight.regenerate")
		: $_("diary.highlight.generate");
	$: buttonLoadingLabel = $_("diary.highlight.generating");
</script>

{#if hasLLMKey}
	<div class="my-4 border border-gray-300 dark:border-gray-600 rounded-lg overflow-hidden">
		<div class="flex flex-col md:flex-row md:justify-between md:items-center gap-4 p-4 bg-gray-50 dark:bg-gray-800 border-b border-gray-300 dark:border-gray-600">
			<button
				type="button"
				class="flex items-center gap-2 text-left hover:opacity-80 transition-opacity"
				on:click={() => (showHighlight = !showHighlight)}
			>
				<span class="text-sm transition-transform duration-200 {showHighlight ? 'rotate-90' : ''}">
					▶
				</span>
				<h3 class="text-lg font-semibold m-0">
					{$_("diary.highlight.label")}
				</h3>
			</button>

			<div class="flex flex-wrap gap-2 items-center">
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
		</div>

		{#if showHighlight}
			<div class="p-4 {highlightJustUpdated ? 'animate-pulse bg-green-50 dark:bg-green-900/20' : ''}">
				{#if isHighlightOutdated && highlightData}
					<div class="flex items-center gap-2 p-3 mb-4 bg-orange-100 dark:bg-orange-900/30 border border-orange-300 dark:border-orange-700 rounded text-sm text-orange-800 dark:text-orange-200">
						<span class="text-xl">⚠️</span>
						{$_("diary.highlight.outdated")}
					</div>
				{/if}

				{#if highlightData && highlightData.highlights.length > 0}
					<div class="p-4 bg-gray-100 dark:bg-gray-900 rounded whitespace-pre-wrap break-words leading-relaxed text-base">
						{@html getHighlightedContent()}
					</div>

					<div class="mt-4 pt-4 border-t border-gray-300 dark:border-gray-600">
						<p class="text-sm text-gray-600 dark:text-gray-400 m-0">
							{$_("diary.highlight.label")}: {highlightData.highlights.length}
						</p>
					</div>
				{:else if highlightStatus === "none"}
					<p class="text-gray-600 dark:text-gray-400 italic m-0">
						{$_("diary.highlight.notAvailable")}
					</p>
				{/if}
			</div>
		{/if}
	</div>
{/if}
