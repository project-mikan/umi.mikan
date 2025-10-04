import { redirect } from "@sveltejs/kit";
import { listEntities } from "$lib/server/entity-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { EntityCategory } from "$lib/grpc/entity/entity_pb";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ url, cookies }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw redirect(302, "/login");
	}

	// カテゴリフィルタパラメータを取得
	const categoryParam = url.searchParams.get("category");
	let category = EntityCategory.NO_CATEGORY; // デフォルトは全て表示

	if (categoryParam === "people") {
		category = EntityCategory.PEOPLE;
	} else if (categoryParam === "noCategory") {
		category = EntityCategory.NO_CATEGORY;
	} else if (categoryParam === "all") {
		category = EntityCategory.NO_CATEGORY; // 0は全て表示の意味
	}

	try {
		const response = await listEntities({
			category: category,
			accessToken: authResult.accessToken,
		});

		return {
			entities: response.entities,
			selectedCategory: categoryParam || "all",
		};
	} catch (err) {
		console.error("Failed to load entities:", err);
		return {
			entities: [],
			selectedCategory: categoryParam || "all",
			error: "Failed to load entities",
		};
	}
};
