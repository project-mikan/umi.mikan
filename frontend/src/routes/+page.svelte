<script lang="ts">
  import { _ } from "svelte-i18n";
  import { beforeNavigate } from "$app/navigation";
  import { browser } from "$app/env";
  import { onMount } from "svelte";
  import "#lib/i18n.js";
  import Button from "#lib/components/atoms/Button.svelte";
  import SaveButton from "#lib/components/atoms/SaveButton.svelte";
  import EditableDiaryCard from "#lib/components/molecules/EditableDiaryCard.svelte";
  import TimeProgressBar from "#lib/components/molecules/TimeProgressBar.svelte";
  import RecentDiaryStreak from "#lib/components/molecules/RecentDiaryStreak.svelte";
  import LatestTrendDisplay from "#lib/components/molecules/LatestTrendDisplay.svelte";
  import CollapsibleSection from "#lib/components/molecules/CollapsibleSection.svelte";
  import type { DiaryDraft } from "#lib/utils/diary-draft.js";
  import type { PageData } from "./$types";

  let title = $derived($_("page.title.home"));

  let { data }: { data: PageData } = $props();

  type DiarySection = "today" | "yesterday" | "dayBeforeYesterday";

  // 各カードの保存中・保存済み・未保存の状態（カードから bind で受け取る）
  let loading = $state<Record<DiarySection, boolean>>({
    today: false,
    yesterday: false,
    dayBeforeYesterday: false,
  });
  let saved = $state<Record<DiarySection, boolean>>({
    today: false,
    yesterday: false,
    dayBeforeYesterday: false,
  });
  let unsaved = $state<Record<DiarySection, boolean>>({
    today: false,
    yesterday: false,
    dayBeforeYesterday: false,
  });
  let editors = $state<Partial<Record<DiarySection, EditableDiaryCard>>>({});
  // 3枚のカードで共有する、日付ごとの書きかけ
  let drafts = $state<Record<string, DiaryDraft>>({});

  // いずれか1つでも未保存の変更があるかチェック
  let hasAnyUnsavedChanges = $derived(
    unsaved.today || unsaved.yesterday || unsaved.dayBeforeYesterday,
  );

  // トグルセクションの開閉状態（localStorageから復元）
  let sideInfoOpen = $state(false);

  // localStorageから開閉状態を復元
  if (browser) {
    const stored = localStorage.getItem("sideInfoOpen");
    if (stored !== null) {
      sideInfoOpen = stored === "true";
    }
  }

  // 開閉状態が変更されたらlocalStorageに保存（$effect はブラウザでのみ実行される）
  $effect(() => {
    localStorage.setItem("sideInfoOpen", String(sideInfoOpen));
  });

  function getMonthlyUrl(): string {
    const now = new Date();
    return `/monthly/${now.getFullYear()}/${now.getMonth() + 1}`;
  }

  // 指定したカードの日記を保存する
  function save(section: DiarySection) {
    editors[section]?.save();
  }

  // ページ遷移前の警告
  beforeNavigate((navigation) => {
    if (hasAnyUnsavedChanges) {
      if (!confirm($_("diary.unsavedChangesWarning"))) {
        navigation.cancel();
      }
    }
  });

  // 表示する保存ボタンを決定（フォーカスまたはスクロール位置に基づく）
  let activeSection: DiarySection = $state("today");
  let todayCard: HTMLElement | undefined = $state();
  let yesterdayCard: HTMLElement | undefined = $state();
  let dayBeforeYesterdayCard: HTMLElement | undefined = $state();

  // フォーカスベースのセクション検出（モバイルキーボード対応）
  let focusedSection: DiarySection | null = null;

  function handleFocusIn(section: DiarySection) {
    focusedSection = section;
    activeSection = section;
  }

  function handleFocusOut() {
    // フォーカスが別のセクションに移る場合があるので、少し遅延してリセット
    setTimeout(() => {
      // フォーカスがどのカードにも無い場合のみスクロールベースに戻す
      const activeEl = document.activeElement;
      if (
        activeEl &&
        (todayCard?.contains(activeEl) ||
          yesterdayCard?.contains(activeEl) ||
          dayBeforeYesterdayCard?.contains(activeEl))
      ) {
        return;
      }
      focusedSection = null;
    }, 100);
  }

  // モバイルキーボード表示時の保存ボタン位置調整
  let saveButtonBottom = $state("5rem"); // デフォルト: bottom-20 (QuickNavigationの上)

  // ブラウザのページ離脱時の警告
  onMount(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasAnyUnsavedChanges) {
        e.preventDefault();
        e.returnValue = "";
      }
    };

    // visualViewport APIを使ってキーボード表示を検出し、保存ボタンの位置を調整
    const updateSaveButtonPosition = () => {
      if (!window.visualViewport) return;
      const viewport = window.visualViewport;
      // キーボードが表示されている場合: layoutViewportの高さとvisualViewportの高さの差
      const keyboardHeight = window.innerHeight - viewport.height;
      const KEYBOARD_THRESHOLD = 100;
      if (keyboardHeight > KEYBOARD_THRESHOLD) {
        // キーボードが表示されている: キーボードの上に配置
        saveButtonBottom = `${keyboardHeight + 8}px`;
      } else {
        // キーボード非表示: QuickNavigationの上に配置
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

    // スクロール位置判定の定数
    const VIEWPORT_CENTER_DIVISOR = 2;

    // スクロール位置を監視して、現在表示中のセクションを判定
    const updateActiveSection = () => {
      // テキストエリアにフォーカスがある場合はスクロールベースの更新をスキップ
      if (focusedSection) return;

      const scrollY = window.scrollY;
      const viewportHeight = window.innerHeight;
      const viewportCenter = scrollY + viewportHeight / VIEWPORT_CENTER_DIVISOR;

      // 各カードの位置を取得
      const todayRect = todayCard?.getBoundingClientRect();
      const yesterdayRect = yesterdayCard?.getBoundingClientRect();
      const dayBeforeYesterdayRect =
        dayBeforeYesterdayCard?.getBoundingClientRect();

      // 画面中央を含むカードを優先的に選択
      // 画面中央がカードの範囲内にある場合、そのカードを選択
      if (
        todayRect &&
        todayRect.top + scrollY <= viewportCenter &&
        todayRect.bottom + scrollY >= viewportCenter
      ) {
        activeSection = "today";
        return;
      }

      if (
        yesterdayRect &&
        yesterdayRect.top + scrollY <= viewportCenter &&
        yesterdayRect.bottom + scrollY >= viewportCenter
      ) {
        activeSection = "yesterday";
        return;
      }

      if (
        dayBeforeYesterdayRect &&
        dayBeforeYesterdayRect.top + scrollY <= viewportCenter &&
        dayBeforeYesterdayRect.bottom + scrollY >= viewportCenter
      ) {
        activeSection = "dayBeforeYesterday";
        return;
      }

      // 画面中央がどのカードにも含まれない場合、画面内で最も近いカードを選択
      const candidates: Array<{
        section: DiarySection;
        distance: number;
      }> = [];

      if (todayRect && todayRect.top < viewportHeight && todayRect.bottom > 0) {
        const todayCenter =
          todayRect.top + scrollY + todayRect.height / VIEWPORT_CENTER_DIVISOR;
        const todayDistance = Math.abs(viewportCenter - todayCenter);
        candidates.push({ section: "today", distance: todayDistance });
      }

      if (
        yesterdayRect &&
        yesterdayRect.top < viewportHeight &&
        yesterdayRect.bottom > 0
      ) {
        const yesterdayCenter =
          yesterdayRect.top +
          scrollY +
          yesterdayRect.height / VIEWPORT_CENTER_DIVISOR;
        const yesterdayDistance = Math.abs(viewportCenter - yesterdayCenter);
        candidates.push({ section: "yesterday", distance: yesterdayDistance });
      }

      if (
        dayBeforeYesterdayRect &&
        dayBeforeYesterdayRect.top < viewportHeight &&
        dayBeforeYesterdayRect.bottom > 0
      ) {
        const dayBeforeYesterdayCenter =
          dayBeforeYesterdayRect.top +
          scrollY +
          dayBeforeYesterdayRect.height / VIEWPORT_CENTER_DIVISOR;
        const dayBeforeYesterdayDistance = Math.abs(
          viewportCenter - dayBeforeYesterdayCenter,
        );
        candidates.push({
          section: "dayBeforeYesterday",
          distance: dayBeforeYesterdayDistance,
        });
      }

      if (candidates.length > 0) {
        candidates.sort((a, b) => a.distance - b.distance);
        activeSection = candidates[0].section;
      }
    };

    // debounce実装（100msの遅延）
    let scrollTimeout: ReturnType<typeof setTimeout> | null = null;
    const handleScroll = () => {
      if (scrollTimeout) clearTimeout(scrollTimeout);
      scrollTimeout = setTimeout(() => {
        updateActiveSection();
      }, 100);
    };

    window.addEventListener("beforeunload", handleBeforeUnload);
    window.addEventListener("scroll", handleScroll, { passive: true });

    // 初期表示時に一度実行
    updateActiveSection();

    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
      window.removeEventListener("scroll", handleScroll);
      if (scrollTimeout) clearTimeout(scrollTimeout);
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
	<!-- 見出し -->
	<div class="flex justify-between items-center mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">{$_("diary.title")}</h1>
	</div>

	<!-- メインコンテンツ: 日記カード -->
	<div class="space-y-6">
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div bind:this={todayCard} onfocusin={() => handleFocusIn("today")} onfocusout={handleFocusOut}>
		<EditableDiaryCard
			bind:this={editors.today}
			bind:loading={loading.today}
			bind:saved={saved.today}
			bind:drafts
			bind:hasUnsavedChanges={unsaved.today}
			title={$_("diary.today")}
			date={data.today.date}
			entry={data.today.entry}
			action="?/saveToday"
			inputId="content"
		/>
		</div>

		<!-- トグル可能なセクション: さいきん -->
		<div>
			<CollapsibleSection title={$_("collapsibleSections.sideInfo")} bind:isOpen={sideInfoOpen}>
				<div class="space-y-4">
					<!-- 進捗バー -->
					<TimeProgressBar />

					<!-- 直近7日 -->
					<RecentDiaryStreak recentDays={data.recentDays} />

					<!-- トレンド分析 -->
					{#if data.autoLatestTrendEnabled}
						<LatestTrendDisplay userName={data.userName} />
					{/if}
				</div>
			</CollapsibleSection>
		</div>

		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div bind:this={yesterdayCard} onfocusin={() => handleFocusIn("yesterday")} onfocusout={handleFocusOut}>
		<EditableDiaryCard
			bind:this={editors.yesterday}
			bind:loading={loading.yesterday}
			bind:saved={saved.yesterday}
			bind:drafts
			bind:hasUnsavedChanges={unsaved.yesterday}
			title={$_("diary.yesterday")}
			date={data.yesterday.date}
			entry={data.yesterday.entry}
			action="?/saveYesterday"
			inputId="yesterday-content"
		/>
		</div>

		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div bind:this={dayBeforeYesterdayCard} onfocusin={() => handleFocusIn("dayBeforeYesterday")} onfocusout={handleFocusOut}>
		<EditableDiaryCard
			bind:this={editors.dayBeforeYesterday}
			bind:loading={loading.dayBeforeYesterday}
			bind:saved={saved.dayBeforeYesterday}
			bind:drafts
			bind:hasUnsavedChanges={unsaved.dayBeforeYesterday}
			title={$_("diary.dayBeforeYesterday")}
			date={data.dayBeforeYesterday.date}
			entry={data.dayBeforeYesterday.entry}
			action="?/saveDayBeforeYesterday"
			inputId="day-before-yesterday-content"
		/>
		</div>
	</div>

	<!-- Fixed Save Button for Mobile (shows only the active section) -->
	<div class="fixed left-0 right-0 p-4 sm:hidden z-10 pointer-events-none transition-[bottom] duration-150" style="bottom: {saveButtonBottom}">
		<div class="max-w-4xl mx-auto flex justify-end pointer-events-auto">
			{#if activeSection === "today"}
				<SaveButton
					type="button"
					loading={loading.today}
					saved={saved.today}
					size="md"
					label={$_("diary.saveTodayDiary")}
					onclick={() => save("today")}
				/>
			{:else if activeSection === "yesterday"}
				<SaveButton
					type="button"
					loading={loading.yesterday}
					saved={saved.yesterday}
					size="md"
					label={$_("diary.saveYesterdayDiary")}
					onclick={() => save("yesterday")}
				/>
			{:else}
				<SaveButton
					type="button"
					loading={loading.dayBeforeYesterday}
					saved={saved.dayBeforeYesterday}
					size="md"
					label={$_("diary.saveDayBeforeYesterdayDiary")}
					onclick={() => save("dayBeforeYesterday")}
				/>
			{/if}
		</div>
	</div>

	<!-- Spacer for fixed button on mobile -->
	<div class="h-32 sm:hidden"></div>
</div>

