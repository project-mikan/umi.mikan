import { describe, expect, it } from "vitest";
import {
  type DiaryDraft,
  createDiaryDraft,
  diaryDateKey,
  hasUnsavedDraft,
  reconcileDiaryDraft,
  reconcileDiaryDraftByDate,
} from "./diary-draft";

describe("diaryDateKey", () => {
  const cases: Array<{
    name: string;
    date: { year: number; month: number; day: number };
    want: string;
  }> = [
    {
      name: "月日が1桁の場合はゼロ埋めする",
      date: { year: 2026, month: 1, day: 2 },
      want: "2026-01-02",
    },
    {
      name: "月日が2桁の場合はそのまま並べる",
      date: { year: 2026, month: 11, day: 30 },
      want: "2026-11-30",
    },
  ];

  for (const tc of cases) {
    it(tc.name, () => {
      expect(diaryDateKey(tc.date)).toBe(tc.want);
    });
  }
});

describe("createDiaryDraft", () => {
  it("サーバーの本文を入力中の本文と基準値の両方に入れる", () => {
    expect(createDiaryDraft("2026-11-02", "本文")).toEqual({
      key: "2026-11-02",
      content: "本文",
      baseline: "本文",
    });
  });
});

describe("hasUnsavedDraft", () => {
  const cases: Array<{ name: string; draft: DiaryDraft; want: boolean }> = [
    {
      name: "入力中の本文が基準値と同じなら未保存の変更はない",
      draft: { key: "k", content: "a", baseline: "a" },
      want: false,
    },
    {
      name: "入力中の本文が基準値と違えば未保存の変更がある",
      draft: { key: "k", content: "ab", baseline: "a" },
      want: true,
    },
  ];

  for (const tc of cases) {
    it(tc.name, () => {
      expect(hasUnsavedDraft(tc.draft)).toBe(tc.want);
    });
  }
});

describe("reconcileDiaryDraft", () => {
  const cases: Array<{
    name: string;
    draft: DiaryDraft;
    key: string;
    serverContent: string;
    want: DiaryDraft;
  }> = [
    {
      // 未保存の変更がないので、サーバーの最新の本文を表示してよい
      name: "未保存の変更がない時はサーバーの本文に置き換える",
      draft: { key: "2026-11-02", content: "古い", baseline: "古い" },
      key: "2026-11-02",
      serverContent: "新しい",
      want: { key: "2026-11-02", content: "新しい", baseline: "新しい" },
    },
    {
      // 別の日記の保存などで再読み込みされても、書きかけを消してはいけない
      name: "未保存の変更がある時は入力中の本文を残し、基準値だけ差し替える",
      draft: { key: "2026-11-02", content: "書きかけ", baseline: "古い" },
      key: "2026-11-02",
      serverContent: "古い",
      want: { key: "2026-11-02", content: "書きかけ", baseline: "古い" },
    },
    {
      // 保存の通信中に打った文字は、保存済みの本文に含まれないので未保存のまま残す
      name: "保存後に追記していた時は追記を残し、未保存として扱う",
      draft: { key: "2026-11-02", content: "保存した本文と追記", baseline: "" },
      key: "2026-11-02",
      serverContent: "保存した本文",
      want: {
        key: "2026-11-02",
        content: "保存した本文と追記",
        baseline: "保存した本文",
      },
    },
    {
      name: "保存した本文がサーバーに反映されたら未保存の変更はなくなる",
      draft: { key: "2026-11-02", content: "保存した本文", baseline: "" },
      key: "2026-11-02",
      serverContent: "保存した本文",
      want: {
        key: "2026-11-02",
        content: "保存した本文",
        baseline: "保存した本文",
      },
    },
    {
      // 遷移前に確認ダイアログを出しているので、前の日の書きかけは捨ててよい
      name: "別の日に切り替わった時は書きかけがあってもサーバーの本文に置き換える",
      draft: { key: "2026-11-01", content: "書きかけ", baseline: "古い" },
      key: "2026-11-02",
      serverContent: "別の日",
      want: { key: "2026-11-02", content: "別の日", baseline: "別の日" },
    },
  ];

  for (const tc of cases) {
    it(tc.name, () => {
      expect(reconcileDiaryDraft(tc.draft, tc.key, tc.serverContent)).toEqual(
        tc.want,
      );
    });
  }
});

describe("reconcileDiaryDraftByDate", () => {
  const drafts: Record<string, DiaryDraft> = {
    "2026-11-01": { key: "2026-11-01", content: "書きかけ", baseline: "古い" },
  };
  const cases: Array<{
    name: string;
    key: string;
    serverContent: string;
    want: DiaryDraft;
  }> = [
    {
      // 日付が変わって「今日」のカードが「昨日」に移っても、その日の書きかけを引き継ぐ
      name: "その日の書きかけがある時は書きかけを残す",
      key: "2026-11-01",
      serverContent: "古い",
      want: { key: "2026-11-01", content: "書きかけ", baseline: "古い" },
    },
    {
      name: "その日の書きかけがない時はサーバーの本文から作る",
      key: "2026-11-02",
      serverContent: "新しい日",
      want: { key: "2026-11-02", content: "新しい日", baseline: "新しい日" },
    },
  ];

  for (const tc of cases) {
    it(tc.name, () => {
      expect(
        reconcileDiaryDraftByDate(drafts, tc.key, tc.serverContent),
      ).toEqual(tc.want);
    });
  }
});
