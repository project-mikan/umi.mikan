import { error } from "@sveltejs/kit";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { exportDiaryEntries } from "$lib/server/diary-api";
import type { RequestHandler } from "./$types";

// GET /api/diary/export?fromYear=2024&fromMonth=4&toYear=2026&toMonth=6
// 指定期間の日記をJSONとしてダウンロードする
export const GET: RequestHandler = async ({ url, cookies }) => {
  const authResult = await ensureValidAccessToken(cookies);

  if (!authResult.isAuthenticated || !authResult.accessToken) {
    throw error(401, "Unauthorized");
  }

  const fromYear = Number(url.searchParams.get("fromYear"));
  const fromMonth = Number(url.searchParams.get("fromMonth"));
  const toYear = Number(url.searchParams.get("toYear"));
  const toMonth = Number(url.searchParams.get("toMonth"));

  // パラメータのバリデーション
  if (!fromYear || !fromMonth || !toYear || !toMonth) {
    throw error(400, "fromYear, fromMonth, toYear, toMonth are required");
  }

  if (
    fromYear < 1900 ||
    fromYear > 2100 ||
    fromMonth < 1 ||
    fromMonth > 12 ||
    toYear < 1900 ||
    toYear > 2100 ||
    toMonth < 1 ||
    toMonth > 12
  ) {
    throw error(400, "Invalid year or month value");
  }

  if (fromYear > toYear || (fromYear === toYear && fromMonth > toMonth)) {
    throw error(
      400,
      "fromYear/fromMonth must be before or equal to toYear/toMonth",
    );
  }

  try {
    const response = await exportDiaryEntries({
      fromYear,
      fromMonth,
      toYear,
      toMonth,
      accessToken: authResult.accessToken,
    });

    // エクスポートデータをJSON形式に変換する
    const exportData = {
      exported_at: new Date().toISOString(),
      period: {
        from: { year: fromYear, month: fromMonth },
        to: { year: toYear, month: toMonth },
      },
      total_count: response.totalCount,
      entries: response.entries.map((entry) => ({
        id: entry.id,
        date: entry.date
          ? {
              year: entry.date.year,
              month: entry.date.month,
              day: entry.date.day,
            }
          : null,
        content: entry.content,
        created_at: entry.createdAt
          ? new Date(Number(entry.createdAt) * 1000).toISOString()
          : null,
        updated_at: entry.updatedAt
          ? new Date(Number(entry.updatedAt) * 1000).toISOString()
          : null,
      })),
    };

    // ファイル名を期間で設定する
    const filename = `diary_export_${fromYear}${String(fromMonth).padStart(2, "0")}-${toYear}${String(toMonth).padStart(2, "0")}.json`;

    return new Response(JSON.stringify(exportData, null, 2), {
      headers: {
        "Content-Type": "application/json; charset=utf-8",
        "Content-Disposition": `attachment; filename="${filename}"`,
      },
    });
  } catch (err) {
    console.error("Export failed:", err);
    throw error(500, "Export failed");
  }
};
