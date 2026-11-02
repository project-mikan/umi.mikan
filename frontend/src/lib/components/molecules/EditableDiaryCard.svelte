<script lang="ts">
  import { _ } from "svelte-i18n";
  import { enhance } from "$app/forms";
  import { untrack } from "svelte";
  import "#lib/i18n.js";
  import SaveButton from "../atoms/SaveButton.svelte";
  import DiaryCard from "./DiaryCard.svelte";
  import FormField from "./FormField.svelte";
  import { createSubmitHandler } from "#lib/utils/form-utils.js";
  import {
    type DiaryDraft,
    createDiaryDraft,
    diaryDateKey,
    hasUnsavedDraft,
    reconcileDiaryDraftByDate,
  } from "#lib/utils/diary-draft.js";
  import type { DiaryEntry, YMD } from "#lib/grpc/diary/diary_pb.js";

  // ホーム画面に並べる、1日分の日記の編集カード
  let {
    title,
    date,
    entry,
    action,
    inputId,
    loading = $bindable(false),
    saved = $bindable(false),
    hasUnsavedChanges = $bindable(false),
    drafts = $bindable({}),
  }: {
    title: string;
    date: YMD;
    entry: DiaryEntry | null | undefined;
    action: string;
    inputId: string;
    loading?: boolean;
    saved?: boolean;
    hasUnsavedChanges?: boolean;
    // 書きかけはカードの位置ではなく日付ごとに持つ。日付が変わってカードの日付がずれても、その日の書きかけを引き継ぐ
    drafts?: Record<string, DiaryDraft>;
  } = $props();

  let formElement: HTMLFormElement | undefined = $state();
  let dateKey = $derived(diaryDateKey(date));
  let serverContent = $derived(entry?.content || "");

  // SSR でも初期表示が正しくなるよう、宣言時にサーバーの本文で初期化する
  untrack(() => {
    drafts[dateKey] = reconcileDiaryDraftByDate(drafts, dateKey, serverContent);
  });
  let draft = $derived(
    drafts[dateKey] ?? createDiaryDraft(dateKey, serverContent),
  );

  // 保存後や他のカードの保存で再読み込みされても、書きかけは消さずに残す
  $effect.pre(() => {
    const key = dateKey;
    const content = serverContent;
    untrack(() => {
      drafts[key] = reconcileDiaryDraftByDate(drafts, key, content);
    });
  });

  let characterCount = $derived(draft.content.length);

  $effect.pre(() => {
    hasUnsavedChanges = hasUnsavedDraft(draft);
  });

  // フォームを送信して日記を保存する（親の保存ボタンからも呼ぶ）
  export function save() {
    formElement?.requestSubmit();
  }
</script>

<DiaryCard {title} entry={entry ?? null} showForm={true} href={`/${dateKey}`}>
	{#snippet editor()}
    <form
  		bind:this={formElement}
  		method="POST"
  		{action}
  		use:enhance={createSubmitHandler(
  			(l) => (loading = l),
  			(s) => (saved = s),
  		)}
  		
  	>
  		<input type="hidden" name="date" value={dateKey} />
  		{#if entry}
  			<input type="hidden" name="id" value={entry.id} />
  		{/if}
  		<FormField
  			type="textarea"
  			label=""
  			id={inputId}
  			name="content"
  			placeholder={$_("diary.placeholder")}
  			rows={8}
  			bind:value={draft.content}
  			onsave={save}
  			onautosave={() => {
  				if (hasUnsavedChanges && !loading) save();
  			}}
  		/>

  		<div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
  			{$_("diary.characterCount", { values: { count: characterCount } })}
  			{#if characterCount >= 1000}
  				<span class="ml-2 text-blue-600 dark:text-blue-400 font-medium">
  					({$_("diary.autoSummaryEligible")})
  				</span>
  			{/if}
  		</div>

  		<div class="sticky bottom-4 flex justify-end hidden sm:flex mt-4 z-10">
  			<SaveButton {loading} {saved} />
  		</div>
  	</form>
  {/snippet}
</DiaryCard>
