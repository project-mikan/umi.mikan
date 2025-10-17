import type { PageServerLoad } from "./$types";
import { error, redirect } from "@sveltejs/kit";
import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import {
	UserService,
	GetPubSubMetricsRequestSchema,
} from "$lib/grpc/user/user_pb";
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
						}
					: {
							totalDailySummaries: 0,
							totalMonthlySummaries: 0,
							pendingDailySummaries: 0,
							pendingMonthlySummaries: 0,
							autoSummaryDailyEnabled: false,
							autoSummaryMonthlyEnabled: false,
						},
			},
		};
	} catch (err) {
		console.error("Failed to load pub/sub metrics:", err);
		throw error(500, "Failed to load metrics data");
	}
};
