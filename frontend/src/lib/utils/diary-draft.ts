// 日記の入力欄の状態。baseline はサーバーに保存済みと分かっている本文
export type DiaryDraft = {
  key: string;
  content: string;
  baseline: string;
};

// 日記を日付で識別するキー（YYYY-MM-DD）を返す
export function diaryDateKey(date: {
  year: number;
  month: number;
  day: number;
}): string {
  return `${date.year}-${String(date.month).padStart(2, "0")}-${String(date.day).padStart(2, "0")}`;
}

// サーバーの本文から入力欄の初期状態を作る
export function createDiaryDraft(
  key: string,
  serverContent: string,
): DiaryDraft {
  return { key, content: serverContent, baseline: serverContent };
}

// 入力中の本文に未保存の変更があるかを返す
export function hasUnsavedDraft(draft: DiaryDraft): boolean {
  return draft.content !== draft.baseline;
}

// ページデータの再読み込み後に入力欄の状態を合わせる。
// 同じ日の書きかけは残し、基準値だけをサーバーの本文に差し替える
export function reconcileDiaryDraft(
  draft: DiaryDraft,
  key: string,
  serverContent: string,
): DiaryDraft {
  if (draft.key !== key || !hasUnsavedDraft(draft)) {
    return createDiaryDraft(key, serverContent);
  }
  return { key, content: draft.content, baseline: serverContent };
}

// 日付ごとの書きかけから指定した日のものを取り出し、サーバーの本文に合わせて返す
export function reconcileDiaryDraftByDate(
  drafts: Record<string, DiaryDraft>,
  key: string,
  serverContent: string,
): DiaryDraft {
  const current = drafts[key] ?? createDiaryDraft(key, serverContent);
  return reconcileDiaryDraft(current, key, serverContent);
}
