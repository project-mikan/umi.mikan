import { error, json } from "@sveltejs/kit";
import { triggerLatestTrend } from "$lib/server/diary-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { RequestHandler } from "./$types";
import {
  AI_NOT_ENABLED_MESSAGE,
  isAiNotEnabledError,
} from "$lib/utils/error-utils";

export const POST: RequestHandler = async ({ cookies }) => {
  const authResult = await ensureValidAccessToken(cookies);

  if (!authResult.isAuthenticated || !authResult.accessToken) {
    throw error(401, "Unauthorized");
  }

  try {
    const triggerResponse = await triggerLatestTrend({
      accessToken: authResult.accessToken,
    });

    if (!triggerResponse.success) {
      return json({
        success: false,
        message:
          triggerResponse.message || "Failed to trigger latest trend analysis",
      });
    }

    return json({
      success: true,
      message:
        triggerResponse.message ||
        "Latest trend analysis generation has been queued",
    });
  } catch (err) {
    console.error("Failed to trigger latest trend:", err);
    console.error("Error details:", {
      message: (err as Error)?.message,
      code: (err as { code?: string })?.code,
      stack: (err as Error)?.stack,
    });

    // AI機能が未有効化（設定画面でオプトインしていない）
    if (isAiNotEnabledError(err)) {
      throw error(400, { message: AI_NOT_ENABLED_MESSAGE });
    }
    if (
      (err as { code?: string })?.code === "NOT_FOUND" ||
      (err as { code?: number })?.code === 5 // gRPC NOT_FOUND code
    ) {
      throw error(404, AI_NOT_ENABLED_MESSAGE);
    }
    if (
      (err as { code?: string })?.code === "PERMISSION_DENIED" ||
      (err as { code?: number })?.code === 7 || // gRPC PERMISSION_DENIED code
      (err as Error)?.message?.includes("production environment")
    ) {
      throw error(
        403,
        "This operation is not available in production environment",
      );
    }
    throw error(500, "Failed to trigger latest trend analysis");
  }
};
