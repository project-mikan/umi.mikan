import { redirect } from "@sveltejs/kit";
import { listEntities } from "$lib/server/entity-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { EntityCategory } from "$lib/grpc/entity/entity_pb";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw redirect(302, "/login");
	}

	try {
		// 全てのエンティティを表示（カテゴリフィルタなし）
		const response = await listEntities({
			category: EntityCategory.NO_CATEGORY,
			allCategories: true,
			accessToken: authResult.accessToken,
		});

		return {
			entities: response.entities,
		};
	} catch (err) {
		console.error("Failed to load entities:", err);
		return {
			entities: [],
			error: "Failed to load entities",
		};
	}
};
