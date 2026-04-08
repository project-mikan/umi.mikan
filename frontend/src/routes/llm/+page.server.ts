import type { Actions, PageServerLoad } from "./$types";
import { error, fail, redirect } from "@sveltejs/kit";
import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import {
  UserService,
  GetPubSubMetricsRequestSchema,
} from "$lib/grpc/user/user_pb";
import { regenerateAllEmbeddings } from "$lib/server/diary-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";

export const load: PageServerLoad = async ({ cookies }) => {
  const authResult = await ensureValidAccessToken(cookies);

  if (!authResult.isAuthenticated || !authResult.accessToken) {
    throw redirect(302, "/login");
  }

  try {
    const transport = createGrpcTransport({
      baseUrl: "http://backend:8080",
    });

    const client = createClient(UserService, transport);

    const request = create(GetPubSubMetricsRequestSchema, {});
    const response = await client.getPubSubMetrics(request, {
      headers: { authorization: `Bearer ${authResult.accessToken}` },
    });

    return {
      metrics: {
        hourlyMetrics: (response.hourlyMetrics || []).map((metric) => ({
          timestamp: Number(metric.timestamp),
          dailySummariesProcessed: metric.dailySummariesProcessed,
          monthlySummariesProcessed: metric.monthlySummariesProcessed,
          dailySummariesFailed: metric.dailySummariesFailed,
          monthlySummariesFailed: metric.monthlySummariesFailed,
          latestTrendsProcessed: metric.latestTrendsProcessed,
          latestTrendsFailed: metric.latestTrendsFailed,
          diaryEmbeddingsProcessed: metric.diaryEmbeddingsProcessed,
          diaryEmbeddingsFailed: metric.diaryEmbeddingsFailed,
          semanticSearchesProcessed: metric.semanticSearchesProcessed,
        })),
        processingTasks: (response.processingTasks || []).map((task) => ({
          taskType: task.taskType,
          date: task.date,
          startedAt: Number(task.startedAt),
        })),
        summary: response.summary
          ? {
              totalDailySummaries: response.summary.totalDailySummaries,
              totalMonthlySummaries: response.summary.totalMonthlySummaries,
              pendingDailySummaries: response.summary.pendingDailySummaries,
              pendingMonthlySummaries: response.summary.pendingMonthlySummaries,
              autoSummaryDailyEnabled: response.summary.autoSummaryDailyEnabled,
              autoSummaryMonthlyEnabled:
                response.summary.autoSummaryMonthlyEnabled,
              autoLatestTrendEnabled: response.summary.autoLatestTrendEnabled,
              latestTrendGeneratedAt: response.summary.latestTrendGeneratedAt,
              semanticSearchEnabled: response.summary.semanticSearchEnabled,
              totalEmbeddings: response.summary.totalEmbeddings,
              totalEmbeddingDiaries: response.summary.totalEmbeddingDiaries,
              pendingEmbeddings: response.summary.pendingEmbeddings,
            }
          : {
              totalDailySummaries: 0,
              totalMonthlySummaries: 0,
              pendingDailySummaries: 0,
              pendingMonthlySummaries: 0,
              autoSummaryDailyEnabled: false,
              autoSummaryMonthlyEnabled: false,
              autoLatestTrendEnabled: false,
              latestTrendGeneratedAt: "",
              semanticSearchEnabled: false,
              totalEmbeddings: 0,
              totalEmbeddingDiaries: 0,
              pendingEmbeddings: 0,
            },
      },
    };
  } catch (err) {
    console.error("Failed to load pub/sub metrics:", err);
    throw error(500, "Failed to load metrics data");
  }
};

export const actions: Actions = {
  regenerateAllEmbeddings: async ({ cookies }) => {
    const authResult = await ensureValidAccessToken(cookies);
    if (!authResult.isAuthenticated || !authResult.accessToken) {
      return fail(401, { error: "unauthorized" });
    }

    try {
      const response = await regenerateAllEmbeddings({
        accessToken: authResult.accessToken,
      });
      return {
        success: true,
        queuedCount: response.queuedCount,
      };
    } catch (err) {
      console.error("Failed to regenerate embeddings:", err);
      return fail(500, { error: "regenerateFailed" });
    }
  },
};
