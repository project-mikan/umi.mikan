import { error, redirect } from "@sveltejs/kit";
import {
  createDiaryEntry,
  createYMD,
  deleteDiaryEntry,
  getDailySummary,
  getDiaryEntry,
  updateDiaryEntry,
  getDiaryEmbeddingStatus,
} from "$lib/server/diary-api";
import { unixToMilliseconds } from "$lib/utils/token-utils";
import { getUserInfo } from "$lib/server/auth-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { getPastSameDates } from "$lib/utils/date-utils";
import type { DiaryEntry } from "$lib/grpc/diary/diary_pb";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({
  params,
  cookies,
  setHeaders,
  depends,
}) => {
  // キャッシュを無効化して常に最新のデータを取得
  setHeaders({
    "cache-control": "no-store, no-cache, must-revalidate, max-age=0",
  });

  // 明示的な依存関係を設定してinvalidateAll()で確実に再読み込み
  depends("diary:entry");

  const authResult = await ensureValidAccessToken(cookies);

  if (!authResult.isAuthenticated || !authResult.accessToken) {
    throw redirect(302, "/login");
  }

  // params.id should be in format YYYY-MM-DD
  const dateMatch = params.id.match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (!dateMatch) {
    throw error(400, "Invalid date format");
  }

  const [, year, month, day] = dateMatch;
  const date = createYMD(
    Number.parseInt(year, 10),
    Number.parseInt(month, 10),
    Number.parseInt(day, 10),
  );

  // 過去の同日の日付を計算
  const pastDates = getPastSameDates({
    year: Number.parseInt(year, 10),
    month: Number.parseInt(month, 10),
    day: Number.parseInt(day, 10),
  });

  const pastDatesArray = [
    pastDates.oneWeekAgo,
    pastDates.oneMonthAgo,
    pastDates.twoMonthsAgo,
    pastDates.sixMonthsAgo,
    pastDates.oneYearAgo,
    pastDates.twoYearsAgo,
    pastDates.threeYearsAgo,
    pastDates.fourYearsAgo,
    pastDates.fiveYearsAgo,
    pastDates.sixYearsAgo,
    pastDates.sevenYearsAgo,
    pastDates.eightYearsAgo,
    pastDates.nineYearsAgo,
    pastDates.tenYearsAgo,
  ];

  const pastEntriesKeys = [
    "oneWeekAgo",
    "oneMonthAgo",
    "twoMonthsAgo",
    "sixMonthsAgo",
    "oneYearAgo",
    "twoYearsAgo",
    "threeYearsAgo",
    "fourYearsAgo",
    "fiveYearsAgo",
    "sixYearsAgo",
    "sevenYearsAgo",
    "eightYearsAgo",
    "nineYearsAgo",
    "tenYearsAgo",
  ] as const;

  // 全プロミスを一斉発火（date は確定済みのため全て並列化可能）
  // メインエントリの NOT_FOUND (code 2) は正常ケースなのでインラインで吸収する
  const entryPromise = getDiaryEntry({
    date,
    accessToken: authResult.accessToken,
  }).catch((err: unknown) => {
    if (err && typeof err === "object" && "code" in err && err.code === 2) {
      return { entry: null };
    }
    throw err;
  });
  const userInfoPromise = getUserInfo({ accessToken: authResult.accessToken });
  const summaryPromise = getDailySummary({
    date,
    accessToken: authResult.accessToken,
  }).catch(() => ({ summary: null }));
  const pastEntryPromises = pastDatesArray.map((pastDate) =>
    getDiaryEntry({
      date: createYMD(pastDate.year, pastDate.month, pastDate.day),
      accessToken: authResult.accessToken as string,
    }).catch(() => ({ entry: null })),
  );

  // 全リクエストを並列で待機（4回のシリアル round-trip → 1回に削減）
  const [[entryResponse, userInfo, summaryResponse], pastEntriesResults] =
    await Promise.all([
      Promise.all([entryPromise, userInfoPromise, summaryPromise]),
      Promise.all(pastEntryPromises),
    ]);

  // 要約を整形
  let dailySummary = null;
  if (summaryResponse.summary) {
    dailySummary = {
      id: summaryResponse.summary.id,
      diaryId: summaryResponse.summary.diaryId,
      date: {
        year: summaryResponse.summary.date?.year || 0,
        month: summaryResponse.summary.date?.month || 0,
        day: summaryResponse.summary.date?.day || 0,
      },
      summary: summaryResponse.summary.summary,
      createdAt: unixToMilliseconds(summaryResponse.summary.createdAt),
      updatedAt: unixToMilliseconds(summaryResponse.summary.updatedAt),
    };
  }

  // RAGインデックス状態を取得（entry.id が必要なため第2フェーズ）
  // Promiseのまま返してストリーミングし、ページ表示をブロックしない
  const semanticSearchEnabled =
    userInfo.llmKeys?.find((k) => k.llmProvider === 1)?.semanticSearchEnabled ??
    false;
  const embeddingStatus =
    entryResponse.entry && semanticSearchEnabled
      ? getDiaryEmbeddingStatus({
          diaryId: entryResponse.entry.id,
          accessToken: authResult.accessToken,
        })
          .then((statusResponse) => ({
            indexed: statusResponse.indexed,
            modelVersion: statusResponse.modelVersion,
            chunkModelVersion: statusResponse.chunkModelVersion,
            createdAt: Number(statusResponse.createdAt),
            updatedAt: Number(statusResponse.updatedAt),
            chunkCount: statusResponse.chunkCount,
            chunkSummaries: statusResponse.chunkSummaries,
          }))
          .catch(() => null)
      : Promise.resolve(null);

  // 過去日記を整形
  const pastEntriesObject = pastEntriesKeys.reduce(
    (acc, key, index) => {
      acc[key] = {
        date: pastDatesArray[index],
        entry: pastEntriesResults[index].entry || null,
      };
      return acc;
    },
    {} as Record<
      (typeof pastEntriesKeys)[number],
      { date: (typeof pastDatesArray)[number]; entry: DiaryEntry | null }
    >,
  );

  const today = new Date();

  return {
    entry: entryResponse.entry || null,
    date,
    pastEntries: pastEntriesObject,
    user: {
      name: userInfo.name,
      email: userInfo.email,
      llmKeys: userInfo.llmKeys || [],
    },
    dailySummary,
    today: {
      year: today.getFullYear(),
      month: today.getMonth() + 1,
      day: today.getDate(),
    },
    semanticSearchEnabled,
    embeddingStatus,
  };
};

export const actions: Actions = {
  save: async ({ request, cookies }) => {
    const authResult = await ensureValidAccessToken(cookies);

    if (!authResult.isAuthenticated || !authResult.accessToken) {
      throw error(401, "Unauthorized");
    }

    const data = await request.formData();
    const content = data.get("content") as string;
    const id = data.get("id") as string;
    const dateStr = data.get("date") as string;

    if (!content || !dateStr) {
      return {
        error: "Content and date are required",
      };
    }

    try {
      // Parse date string directly to avoid timezone issues
      const dateMatch = dateStr.match(/^(\d{4})-(\d{2})-(\d{2})$/);
      if (!dateMatch) {
        return {
          error: "Invalid date format",
        };
      }
      const [, year, month, day] = dateMatch;
      const ymd = createYMD(
        Number.parseInt(year, 10),
        Number.parseInt(month, 10),
        Number.parseInt(day, 10),
      );

      if (id) {
        // 既存の日記を更新
        await updateDiaryEntry({
          id,
          title: "",
          content,
          date: ymd,
          accessToken: authResult.accessToken,
        });
      } else {
        // 新しい日記を作成
        await createDiaryEntry({
          content,
          date: ymd,
          accessToken: authResult.accessToken,
        });
      }
    } catch (err) {
      if (err instanceof Response) {
        throw err;
      }
      console.error("Failed to save diary entry:", err);
      return {
        error: "Failed to save diary entry",
      };
    }

    return {
      success: true,
    };
  },

  delete: async ({ params, cookies }) => {
    const authResult = await ensureValidAccessToken(cookies);

    if (!authResult.isAuthenticated || !authResult.accessToken) {
      throw error(401, "Unauthorized");
    }

    try {
      // First, get the current entry to get the ID
      const dateMatch = params.id.match(/^(\d{4})-(\d{2})-(\d{2})$/);
      if (!dateMatch) {
        throw error(400, "Invalid date format");
      }

      const [, year, month, day] = dateMatch;
      let currentResponse: Awaited<ReturnType<typeof getDiaryEntry>>;
      try {
        currentResponse = await getDiaryEntry({
          date: createYMD(
            Number.parseInt(year, 10),
            Number.parseInt(month, 10),
            Number.parseInt(day, 10),
          ),
          accessToken: authResult.accessToken,
        });
      } catch (getDiaryErr) {
        // Handle gRPC NOT_FOUND error (code 2) - diary entry doesn't exist
        if (
          getDiaryErr &&
          typeof getDiaryErr === "object" &&
          "code" in getDiaryErr &&
          getDiaryErr.code === 2
        ) {
          return {
            error: "Diary entry not found",
          };
        }
        throw getDiaryErr;
      }

      if (!currentResponse.entry) {
        return {
          error: "Diary entry not found",
        };
      }

      await deleteDiaryEntry({
        id: currentResponse.entry.id,
        accessToken: authResult.accessToken,
      });
    } catch (err) {
      if (err instanceof Response) {
        throw err;
      }
      console.error("Failed to delete diary entry:", err);
      return {
        error: "Failed to delete diary entry",
      };
    }

    throw redirect(303, "/");
  },
};
