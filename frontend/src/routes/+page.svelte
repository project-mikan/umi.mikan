<script lang="ts">
  import { _ } from "svelte-i18n";
  import { enhance } from "$app/forms";
  import { goto } from "$app/navigation";
  import { beforeNavigate } from "$app/navigation";
  import { page } from "$app/state";
  import { onMount, untrack } from "svelte";
  import "#lib/i18n.js";
  import Button from "#lib/components/atoms/Button.svelte";
  import SaveButton from "#lib/components/atoms/SaveButton.svelte";
  import DiaryCard from "#lib/components/molecules/DiaryCard.svelte";
  import FormField from "#lib/components/molecules/FormField.svelte";
  import TimeProgressBar from "#lib/components/molecules/TimeProgressBar.svelte";
  import RecentDiaryStreak from "#lib/components/molecules/RecentDiaryStreak.svelte";
  import LatestTrendDisplay from "#lib/components/molecules/LatestTrendDisplay.svelte";
  import CollapsibleSection from "#lib/components/molecules/CollapsibleSection.svelte";
  import { createSubmitHandler } from "#lib/utils/form-utils.js";
  import type { DiaryEntry, YMD } from "#lib/grpc/diary/diary_pb.js";
  import type { PageData } from "./$types";

  let title = $derived($_("page.title.home"));

  let { data }: { data: PageData } = $props();

  // 入力中の本文。data が変わる（保存後の再読み込み・日付の切り替え）とサーバーの値に戻る
  let todayContent = $derived(data.today.entry?.content || "");
  let yesterdayContent = $derived(data.yesterday.entry?.content || "");
  let dayBeforeYesterdayContent = $derived(
    data.dayBeforeYesterday.entry?.content || "",
  );

  let formElement: HTMLFormElement | undefined = $state();
  let yesterdayFormElement: HTMLFormElement | undefined = $state();
  let dayBeforeYesterdayFormElement: HTMLFormElement | undefined = $state();
  let todayLoading = $state(false);
  let yesterdayLoading = $state(false);
  let dayBeforeLoading = $state(false);
  let todaySaved = $state(false);
  let yesterdaySaved = $state(false);
  let dayBeforeSaved = $state(false);

  // Character count calculations
  let todayCharacterCount = $derived(todayContent ? todayContent.length : 0);
  let yesterdayCharacterCount = $derived(
    yesterdayContent ? yesterdayContent.length : 0,
  );
  let dayBeforeYesterdayCharacterCount = $derived(
    dayBeforeYesterdayContent ? dayBeforeYesterdayContent.length : 0,
  );

  // dataの一意性を判定するためのID（日付の組み合わせ）
  function dataIdOf(d: PageData): string {
    return `${d.today.date.year}-${d.today.date.month}-${d.today.date.day}`;
  }

  // 未保存状態の管理（初期値は表示時の日記。SSRでも正しい値になるようここで設定する）
  let initialTodayContent = $state(
    untrack(() => data.today.entry?.content || ""),
  );
  let initialYesterdayContent = $state(
    untrack(() => data.yesterday.entry?.content || ""),
  );
  let initialDayBeforeYesterdayContent = $state(
    untrack(() => data.dayBeforeYesterday.entry?.content || ""),
  );
  let allowNavigation = $state(false);

  // 前回のdataのIDを保持して、日付が切り替わったことを検出する
  let previousDataId = untrack(() => dataIdOf(data));

  // 日付が切り替わった時（dataのIDが変わった時）だけ初期コンテンツをリセットする
  $effect.pre(() => {
    const currentDataId = dataIdOf(data);
    const today = data.today.entry?.content || "";
    const yesterday = data.yesterday.entry?.content || "";
    const dayBeforeYesterday = data.dayBeforeYesterday.entry?.content || "";
    untrack(() => {
      if (currentDataId === previousDataId) return;
      previousDataId = currentDataId;
      initialTodayContent = today;
      initialYesterdayContent = yesterday;
      initialDayBeforeYesterdayContent = dayBeforeYesterday;
      todayContent = today;
      yesterdayContent = yesterday;
      dayBeforeYesterdayContent = dayBeforeYesterday;
      // 新しいページではallowNavigationをリセット
      allowNavigation = false;
    });
  });

  // 各日記の未保存状態を監視
  let todayHasUnsavedChanges = $derived(
    todayContent !== initialTodayContent && !allowNavigation,
  );
  let yesterdayHasUnsavedChanges = $derived(
    yesterdayContent !== initialYesterdayContent && !allowNavigation,
  );
  let dayBeforeYesterdayHasUnsavedChanges = $derived(
    dayBeforeYesterdayContent !== initialDayBeforeYesterdayContent &&
      !allowNavigation,
  );

  // いずれか1つでも未保存の変更があるかチェック
  let hasAnyUnsavedChanges = $derived(
    todayHasUnsavedChanges ||
      yesterdayHasUnsavedChanges ||
      dayBeforeYesterdayHasUnsavedChanges,
  );

  // トグルセクションの開閉状態（localStorageから復元）
  let sideInfoOpen = $state(false);

  // localStorageから開閉状態を復元
  if (typeof window !== "undefined") {
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

  function formatDateStr(ymd: YMD): string {
    return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
  }

  function viewEntry(entry: DiaryEntry) {
    const date = entry.date;
    if (date) {
      const dateStr = formatDateStr(date);
      goto(`/${dateStr}`);
    }
  }

  function handleSave() {
    formElement?.requestSubmit();
  }

  function handleYesterdaySave() {
    yesterdayFormElement?.requestSubmit();
  }

  function handleDayBeforeYesterdaySave() {
    dayBeforeYesterdayFormElement?.requestSubmit();
  }

  // ページ遷移前の警告
  beforeNavigate((navigation) => {
    if (hasAnyUnsavedChanges && !allowNavigation) {
      if (!confirm($_("diary.unsavedChangesWarning"))) {
        navigation.cancel();
      }
    }
  });

  // 表示する保存ボタンを決定（フォーカスまたはスクロール位置に基づく）
  let activeSection: "today" | "yesterday" | "dayBeforeYesterday" =
    $state("today");
  let todayCard: HTMLElement | undefined = $state();
  let yesterdayCard: HTMLElement | undefined = $state();
  let dayBeforeYesterdayCard: HTMLElement | undefined = $state();

  // フォーカスベースのセクション検出（モバイルキーボード対応）
  let focusedSection: "today" | "yesterday" | "dayBeforeYesterday" | null =
    null;

  function handleFocusIn(
    section: "today" | "yesterday" | "dayBeforeYesterday",
  ) {
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
        section: "today" | "yesterday" | "dayBeforeYesterday";
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
		<DiaryCard
			title={$_("diary.today")}
			entry={data.today.entry}
			showForm={true}
			href={`/${formatDateStr(data.today.date)}`}
		>
			<form
				bind:this={formElement}
				method="POST"
				action="?/saveToday"
use:enhance={createSubmitHandler(
	(loading) => todayLoading = loading,
	(saved) => {
		todaySaved = saved;
		if (saved) {
			// 保存成功時に初期コンテンツを更新
			initialTodayContent = todayContent;
			// 個別のフラグでallowNavigationを制御せず、hasUnsavedChangesの再計算に任せる
		}
	}
)}
				slot="form"
			>
				<input
					type="hidden"
					name="date"
					value={formatDateStr(data.today.date)}
				/>
				{#if data.today.entry}
					<input type="hidden" name="id" value={data.today.entry.id} />
				{/if}
				<FormField
					type="textarea"
					label=""
					id="content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					bind:value={todayContent}
					on:save={handleSave}
					on:autosave={() => { if (todayHasUnsavedChanges && !todayLoading) handleSave(); }}
				/>

				<!-- Character count display -->
				<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{$_("diary.characterCount", { values: { count: todayCharacterCount } })}
					{#if todayCharacterCount >= 1000}
						<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
							({$_("diary.autoSummaryEligible")})
						</span>
					{/if}
				</div>

				<div class="sticky bottom-4 flex justify-end hidden sm:flex mt-4 z-10">
					<SaveButton loading={todayLoading} saved={todaySaved} />
				</div>
			</form>
		</DiaryCard>
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
					{#if page.data.autoLatestTrendEnabled}
						<LatestTrendDisplay userName={page.data.userName} />
					{/if}
				</div>
			</CollapsibleSection>
		</div>

		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div bind:this={yesterdayCard} onfocusin={() => handleFocusIn("yesterday")} onfocusout={handleFocusOut}>
		<DiaryCard
			title={$_("diary.yesterday")}
			entry={data.yesterday.entry}
			showForm={true}
			href={`/${formatDateStr(data.yesterday.date)}`}
		>
			<form
				bind:this={yesterdayFormElement}
				method="POST"
				action="?/saveYesterday"
use:enhance={createSubmitHandler(
	(loading) => yesterdayLoading = loading,
	(saved) => {
		yesterdaySaved = saved;
		if (saved) {
			// 保存成功時に初期コンテンツを更新
			initialYesterdayContent = yesterdayContent;
		}
	}
)}
				slot="form"
			>
				<input
					type="hidden"
					name="date"
					value={formatDateStr(data.yesterday.date)}
				/>
				{#if data.yesterday.entry}
					<input type="hidden" name="id" value={data.yesterday.entry.id} />
				{/if}
				<FormField
					type="textarea"
					label=""
					id="yesterday-content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					bind:value={yesterdayContent}
					on:save={handleYesterdaySave}
					on:autosave={() => { if (yesterdayHasUnsavedChanges && !yesterdayLoading) handleYesterdaySave(); }}
				/>

				<!-- Character count display -->
				<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{$_("diary.characterCount", { values: { count: yesterdayCharacterCount } })}
					{#if yesterdayCharacterCount >= 1000}
						<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
							({$_("diary.autoSummaryEligible")})
						</span>
					{/if}
				</div>

				<div class="sticky bottom-4 flex justify-end hidden sm:flex mt-4 z-10">
					<SaveButton loading={yesterdayLoading} saved={yesterdaySaved} />
				</div>
			</form>
		</DiaryCard>
		</div>

		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div bind:this={dayBeforeYesterdayCard} onfocusin={() => handleFocusIn("dayBeforeYesterday")} onfocusout={handleFocusOut}>
		<DiaryCard
			title={$_("diary.dayBeforeYesterday")}
			entry={data.dayBeforeYesterday.entry}
			showForm={true}
			href={`/${formatDateStr(data.dayBeforeYesterday.date)}`}
		>
			<form
				bind:this={dayBeforeYesterdayFormElement}
				method="POST"
				action="?/saveDayBeforeYesterday"
use:enhance={createSubmitHandler(
	(loading) => dayBeforeLoading = loading,
	(saved) => {
		dayBeforeSaved = saved;
		if (saved) {
			// 保存成功時に初期コンテンツを更新
			initialDayBeforeYesterdayContent = dayBeforeYesterdayContent;
		}
	}
)}
				slot="form"
			>
				<input
					type="hidden"
					name="date"
					value={formatDateStr(data.dayBeforeYesterday.date)}
				/>
				{#if data.dayBeforeYesterday.entry}
					<input
						type="hidden"
						name="id"
						value={data.dayBeforeYesterday.entry.id}
					/>
				{/if}
				<FormField
					type="textarea"
					label=""
					id="day-before-yesterday-content"
					name="content"
					placeholder={$_("diary.placeholder")}
					rows={8}
					bind:value={dayBeforeYesterdayContent}
					on:save={handleDayBeforeYesterdaySave}
					on:autosave={() => { if (dayBeforeYesterdayHasUnsavedChanges && !dayBeforeLoading) handleDayBeforeYesterdaySave(); }}
				/>

				<!-- Character count display -->
				<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
					{$_("diary.characterCount", { values: { count: dayBeforeYesterdayCharacterCount } })}
					{#if dayBeforeYesterdayCharacterCount >= 1000}
						<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
							({$_("diary.autoSummaryEligible")})
						</span>
					{/if}
				</div>

				<div class="sticky bottom-4 flex justify-end hidden sm:flex mt-4 z-10">
					<SaveButton loading={dayBeforeLoading} saved={dayBeforeSaved} />
				</div>
			</form>
		</DiaryCard>
		</div>
	</div>

	<!-- Fixed Save Button for Mobile (shows only the active section) -->
	<div class="fixed left-0 right-0 p-4 sm:hidden z-10 pointer-events-none transition-[bottom] duration-150" style="bottom: {saveButtonBottom}">
		<div class="max-w-4xl mx-auto flex justify-end pointer-events-auto">
			{#if activeSection === "today"}
				<SaveButton
					type="button"
					loading={todayLoading}
					saved={todaySaved}
					size="md"
					label={$_("diary.saveTodayDiary")}
					on:click={handleSave}
				/>
			{:else if activeSection === "yesterday"}
				<SaveButton
					type="button"
					loading={yesterdayLoading}
					saved={yesterdaySaved}
					size="md"
					label={$_("diary.saveYesterdayDiary")}
					on:click={handleYesterdaySave}
				/>
			{:else}
				<SaveButton
					type="button"
					loading={dayBeforeLoading}
					saved={dayBeforeSaved}
					size="md"
					label={$_("diary.saveDayBeforeYesterdayDiary")}
					on:click={handleDayBeforeYesterdaySave}
				/>
			{/if}
		</div>
	</div>

	<!-- Spacer for fixed button on mobile -->
	<div class="h-32 sm:hidden"></div>
</div>

