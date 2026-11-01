<script lang="ts">
  import { _, locale } from "svelte-i18n";
  import { enhance } from "$app/forms";
  import { goto, refreshAll } from "$app/navigation";
  import { beforeNavigate } from "$app/navigation";
  import { page } from "$app/state";
  import { onMount, untrack } from "svelte";
  import "#lib/i18n.js";
  import Button from "#lib/components/atoms/Button.svelte";
  import SaveButton from "#lib/components/atoms/SaveButton.svelte";
  import DiaryCard from "#lib/components/molecules/DiaryCard.svelte";
  import DiaryNavigation from "#lib/components/molecules/DiaryNavigation.svelte";
  import FormField from "#lib/components/molecules/FormField.svelte";
  import HighlightDisplay from "#lib/components/molecules/HighlightDisplay.svelte";
  import Modal from "#lib/components/molecules/Modal.svelte";
  import PastEntriesLinks from "#lib/components/molecules/PastEntriesLinks.svelte";
  import DiaryChunkTimeline from "#lib/components/molecules/DiaryChunkTimeline.svelte";
  import { getDayOfWeekKey } from "#lib/utils/date-utils.js";
  import { createSubmitHandler } from "#lib/utils/form-utils.js";
  import type { HighlightData } from "#lib/types/highlight.js";
  import type { ActionData, PageData } from "./$types";

  let { data, form }: { data: PageData; form: ActionData } = $props();

  // 曜日付きの日付表示（$_ を呼び出し時に読むので、言語切り替えにも追従する）
  function _formatDate(ymd: {
    year: number;
    month: number;
    day: number;
  }): string {
    const dayOfWeekKey = getDayOfWeekKey(ymd);
    const dayOfWeek = $_(`date.dayOfWeek.${dayOfWeekKey}`);
    return $_("date.format.yearMonthDayWithDayOfWeek", {
      values: {
        year: ymd.year,
        month: ymd.month,
        day: ymd.day,
        dayOfWeek: dayOfWeek,
      },
    });
  }

  let title = $derived(
    $_("page.title.individual", {
      values: {
        date: _formatDate(data.date),
      },
    }),
  );

  // 入力中の本文。data が変わる（保存後の再読み込み・別の日への遷移）とサーバーの値に戻る
  let content = $derived(data.entry?.content || "");
  let formElement: HTMLFormElement | undefined = $state();
  let _showDeleteConfirm = $state(false);
  let loading = $state(false);
  let saved = $state(false);
  let isHighlightOutdated = $state(false); // ハイライトが古いかどうか

  // ハイライトデータ
  let highlightData = $state<HighlightData | null>(null);
  let highlightVisible = $state(true); // ハイライトの表示・非表示状態

  // 検索ハイライト（URLの?searchパラメータから取得）
  let searchKeyword = $derived(page.url.searchParams.get("search") ?? "");

  // 検索ハイライトをクリア
  function _clearSearchHighlight() {
    goto(page.url.pathname, { replace: true });
  }

  // Textareaに渡すハイライトデータ（表示・非表示を反映）
  // 配列の参照を毎回変更してTextareaのリアクティビティを確実にトリガー
  let displayedHighlights = $derived(
    highlightVisible && highlightData ? [...highlightData.highlights] : [],
  );

  // 日記の一意性を判定するためのID（entryがない日は日付で判定）
  function entryIdOf(d: PageData): string {
    return d.entry?.id || `${d.date.year}-${d.date.month}-${d.date.day}`;
  }

  // 未保存状態の管理（初期値は表示時の日記。SSRでも正しい値になるようここで設定する）
  let initialContent = $state(untrack(() => data.entry?.content || ""));
  let allowNavigation = $state(false);

  // 前回のentryのIDを保持して、別の日記に遷移したことを検出する
  let previousEntryId = untrack(() => entryIdOf(data));

  // コンテンツの変更を監視して未保存状態を更新
  let hasUnsavedChanges = $derived(
    content !== initialContent && !allowNavigation,
  );

  // 日付判定（当日・未来日）
  let dateFlags = $derived.by(() => {
    const now = new Date();
    const currentDate = new Date(
      data.date.year,
      data.date.month - 1,
      data.date.day,
    );
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    return {
      isToday: currentDate.getTime() === today.getTime(),
      isFutureDate: currentDate.getTime() > today.getTime(),
    };
  });
  let isToday = $derived(dateFlags.isToday);
  let isFutureDate = $derived(dateFlags.isFutureDate);

  // ハイライトが古いか（日記の updatedAt > ハイライトの updatedAt）を判定し、古ければ非表示にする
  $effect.pre(() => {
    const entry = data.entry;
    const highlight = highlightData;
    untrack(() => {
      if (!highlight || !entry) {
        isHighlightOutdated = false;
        return;
      }
      // 日記エントリは秒単位、ハイライトはミリ秒単位なので統一
      const diaryUpdatedAt = Number(entry.updatedAt) * 1000; // 秒 → ミリ秒
      const highlightUpdatedAt = Number(highlight.updatedAt); // 既にミリ秒

      // ハイライトが日記よりも古い場合
      const isOutdated = diaryUpdatedAt > highlightUpdatedAt;

      // 古くなった場合は非表示にする
      if (isOutdated && !isHighlightOutdated) {
        isHighlightOutdated = true;
        highlightVisible = false;
      } else if (!isOutdated) {
        isHighlightOutdated = false;
      }
    });
  });

  // Character count calculation
  let characterCount = $derived(content ? content.length : 0);

  // 別の日記に遷移した時（entryのIDが変わった時）だけ未保存状態を初期化する
  $effect.pre(() => {
    const currentEntryId = entryIdOf(data);
    const entryContent = data.entry?.content || "";
    untrack(() => {
      if (currentEntryId === previousEntryId) return;
      previousEntryId = currentEntryId;
      initialContent = entryContent;
      content = entryContent;
      // 新しいページではallowNavigationをリセット
      allowNavigation = false;
    });
  });

  function handleHighlightUpdated(event: CustomEvent) {
    const newHighlight = event.detail.highlight;
    highlightData = newHighlight;
    // ハイライトが更新されたら古くないとマークして表示する
    isHighlightOutdated = false;
    highlightVisible = true;
  }

  function handleHighlightVisibilityChanged(event: CustomEvent) {
    // ハイライトの表示・非表示状態を更新
    highlightVisible = event.detail.visible;
  }

  function _formatDateStr(ymd: {
    year: number;
    month: number;
    day: number;
  }): string {
    return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
  }

  function _goBack() {
    goto("/");
  }

  function _goToMonthly() {
    const year = data.date.year;
    const month = String(data.date.month).padStart(2, "0");
    goto(`/monthly/${year}/${month}`);
  }

  function _handleSave() {
    formElement?.requestSubmit();
  }

  function _confirmDelete() {
    _showDeleteConfirm = true;
  }

  function _cancelDelete() {
    _showDeleteConfirm = false;
  }

  function _handleDelete() {
    // 削除時は遷移を許可
    allowNavigation = true;
    const form = document.createElement("form");
    form.method = "POST";
    form.action = "?/delete";
    document.body.appendChild(form);
    form.submit();
  }

  // ページ遷移前の警告
  beforeNavigate((navigation) => {
    if (hasUnsavedChanges && !allowNavigation) {
      if (!confirm($_("diary.unsavedChangesWarning"))) {
        navigation.cancel();
      }
    }
  });

  // モバイルキーボード表示時の保存ボタン位置調整
  let saveButtonBottom = $state("5rem"); // デフォルト: bottom-20 (QuickNavigationの上)

  // ブラウザのページ離脱時の警告
  onMount(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasUnsavedChanges) {
        e.preventDefault();
        e.returnValue = "";
      }
    };

    // visualViewport APIを使ってキーボード表示を検出し、保存ボタンの位置を調整
    const updateSaveButtonPosition = () => {
      if (!window.visualViewport) return;
      const viewport = window.visualViewport;
      const keyboardHeight = window.innerHeight - viewport.height;
      const KEYBOARD_THRESHOLD = 100;
      if (keyboardHeight > KEYBOARD_THRESHOLD) {
        saveButtonBottom = `${keyboardHeight + 8}px`;
      } else {
        saveButtonBottom = "5rem";
      }
    };

    if (window.visualViewport) {
      window.visualViewport.addEventListener(
        "resize",
        updateSaveButtonPosition,
      );
      window.visualViewport.addEventListener(
        "scroll",
        updateSaveButtonPosition,
      );
    }

    window.addEventListener("beforeunload", handleBeforeUnload);

    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
      if (window.visualViewport) {
        window.visualViewport.removeEventListener(
          "resize",
          updateSaveButtonPosition,
        );
        window.visualViewport.removeEventListener(
          "scroll",
          updateSaveButtonPosition,
        );
      }
    };
  });
</script>

<svelte:head>
	<title>{title}</title>
</svelte:head>

<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">{$_("diary.title")}</h1>
		<div class="flex gap-2">
			<button
				onclick={_goToMonthly}
				class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium"
			>
				{$_("diary.viewThisMonth")}
			</button>
		</div>
	</div>

	<div class="space-y-6">
		<DiaryNavigation currentDate={data.date} />


		<!-- チャンクタイムライン（embeddingがある場合のみ表示） -->
		{#if data.semanticSearchEnabled && data.entry}
			<DiaryChunkTimeline
				embeddingStatus={data.embeddingStatus}
			/>
		{/if}

		<DiaryCard
			title=""
			entry={data.entry}
			showForm={true}
		>
			<form
				bind:this={formElement}
				method="POST"
				action="?/save"
use:enhance={createSubmitHandler(
	(l) => loading = l,
	async (s) => {
		saved = s;
		if (s) {
			// 保存成功時にページデータを再読み込み（updatedAtを更新）
			await refreshAll();
			// 保存成功時に初期コンテンツを更新
			initialContent = content;
			// hasUnsavedChangesの再計算に任せる
			// ハイライトの古い/新しいの判定はリアクティブステートメントに任せる
		}
	}
)}
				slot="form"
			>
				<input type="hidden" name="date" value={_formatDateStr(data.date)} />
				{#if data.entry}
					<input type="hidden" name="id" value={data.entry.id} />
				{/if}
				<!-- Highlight controls -->
				{#if data.entry && characterCount >= 500}
					<HighlightDisplay
						diaryId={data.entry.id}
						{isHighlightOutdated}
						diaryUpdatedAt={Number(data.entry.updatedAt)}
						on:highlightUpdated={handleHighlightUpdated}
						on:highlightVisibilityChanged={handleHighlightVisibilityChanged}
					/>
				{/if}

				<!-- 検索ハイライトバナー -->
				{#if searchKeyword}
					<div class="flex items-center gap-2 px-3 py-2 mb-2 bg-orange-50 dark:bg-orange-900/20 border border-orange-200 dark:border-orange-700 rounded-md text-sm">
						<svg class="w-4 h-4 text-orange-500 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
							<path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
						</svg>
						<span class="text-orange-800 dark:text-orange-200 flex-1">
							{$_("diary.searchHighlightBanner", { values: { keyword: searchKeyword } })}
						</span>
						<button
							type="button"
							onclick={_clearSearchHighlight}
							class="text-orange-600 dark:text-orange-400 hover:text-orange-800 dark:hover:text-orange-200 font-medium"
						>
							{$_("diary.searchHighlightClear")}
						</button>
					</div>
				{/if}

				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					diaryHighlights={displayedHighlights}
					{searchKeyword}
					bind:value={content}
					on:save={_handleSave}
					on:autosave={() => { if (hasUnsavedChanges && !loading) _handleSave(); }}
				/>

				{#if form?.error}
					<div class="mt-2 text-sm text-red-600 dark:text-red-400">
						{form.error}
					</div>
				{/if}

				<!-- Character count display -->
				<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{$_("diary.characterCount", { values: { count: characterCount } })}
					{#if characterCount >= 1000}
						<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
							({$_("diary.autoSummaryEligible")})
						</span>
					{/if}
					{#if data.entry?.updatedAt}
						<span class="ml-4">
							{$_("common.updatedAt")}: {new Date(Number(data.entry.updatedAt) * 1000).toLocaleString()}
						</span>
					{/if}
				</div>



				<div class="flex justify-between">
					<div>
						{#if data.entry}
							<Button
								type="button"
								variant="danger"
								size="md"
								on:click={_confirmDelete}
							>
								{$_("diary.delete")}
							</Button>
						{/if}
					</div>
				</div>

				<div class="sticky bottom-4 flex justify-end hidden sm:flex mt-4 z-10">
					<SaveButton {loading} {saved} />
				</div>
			</form>
		</DiaryCard>

		<PastEntriesLinks pastEntries={data.pastEntries} />
	</div>

	<!-- Fixed Save Button for Mobile -->
	<div class="fixed left-0 right-0 p-4 sm:hidden z-10 pointer-events-none transition-[bottom] duration-150" style="bottom: {saveButtonBottom}">
		<div class="max-w-4xl mx-auto flex justify-end pointer-events-auto">
			<SaveButton
				type="button"
				{loading}
				{saved}
				size="md"
				on:click={_handleSave}
			/>
		</div>
	</div>

	<!-- Spacer for fixed button on mobile -->
	<div class="h-32 sm:hidden"></div>
</div>

<Modal
	isOpen={_showDeleteConfirm}
	title={$_("edit.deleteConfirm")}
	confirmText={$_("diary.delete")}
	cancelText={$_("diary.cancel")}
	variant="danger"
	onConfirm={_handleDelete}
	onCancel={_cancelDelete}
>
	<p class="text-sm text-gray-500 dark:text-gray-400">
		{$_("edit.deleteMessage")}
	</p>
</Modal>


