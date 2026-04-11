import { error, json } from "@sveltejs/kit";
import { createYM, generateMonthlySummary } from "$lib/server/diary-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { unixToMilliseconds } from "$lib/utils/token-utils";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({ cookies, request }) => {
  const authResult = await ensureValidAccessToken(cookies);

  if (!authResult.isAuthenticated || !authResult.accessToken) {
    throw error(401, "Unauthorized");
  }

  let body: { year: number; month: number };
  try {
    body = await request.json();
  } catch {
    throw error(400, "Invalid JSON body");
  }

  const { year, month } = body;

  if (
    typeof year !== "number" ||
    typeof month !== "number" ||
    month < 1 ||
    month > 12
  ) {
    throw error(400, "Invalid year or month");
  }

  try {
    const summaryResponse = await generateMonthlySummary({
      month: createYM(year, month),
      accessToken: authResult.accessToken,
    });

    if (!summaryResponse.summary) {
      throw error(500, "Failed to generate summary");
    }

    return json({
      summary: {
        id: summaryResponse.summary.id,
        month: {
          year: summaryResponse.summary.month?.year,
          month: summaryResponse.summary.month?.month,
        },
        summary: summaryResponse.summary.summary,
        createdAt: unixToMilliseconds(summaryResponse.summary.createdAt || 0),
        updatedAt: unixToMilliseconds(summaryResponse.summary.updatedAt || 0),
        modelVersion: summaryResponse.summary.modelVersion,
      },
    });
  } catch (err) {
    if (err instanceof Response) {
      throw err;
    }
    if ((err as Error)?.message?.includes("API key")) {
      throw error(400, { message: "Gemini API key not configured" });
    }
    if (
      (err as Error)?.message?.includes("only allowed for past months") ||
      (err as { code?: number })?.code === 9 // gRPC FAILED_PRECONDITION
    ) {
      throw error(400, {
        message: "Monthly summary generation is only allowed for past months",
      });
    }
    // NOT_FOUND は正常ケース（対象月に日記エントリが存在しない）
    if (
      (err as { code?: number })?.code === 5 ||
      (err as Error)?.message?.toLowerCase().includes("not found") ||
      (err as Error)?.message?.includes("no diary entries")
    ) {
      throw error(404, "No diary entries found for the specified month");
    }
    console.error("Failed to generate monthly summary:", err);
    throw error(500, "Failed to generate monthly summary");
  }
};
