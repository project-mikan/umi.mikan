import { json } from "@sveltejs/kit";
import { searchEntities } from "$lib/server/entity-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({ url, cookies }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		return json({ error: "Unauthorized" }, { status: 401 });
	}

	const query = url.searchParams.get("q") || "";

	try {
		const response = await searchEntities({
			query,
			accessToken: authResult.accessToken,
		});

		// BigIntをStringに変換してJSONシリアライズ可能にする
		const entities = response.entities.map((entity) => ({
			...entity,
			createdAt: entity.createdAt.toString(),
			updatedAt: entity.updatedAt.toString(),
			aliases: entity.aliases.map((alias) => ({
				...alias,
				createdAt: alias.createdAt.toString(),
				updatedAt: alias.updatedAt.toString(),
			})),
		}));

		return json({ entities });
	} catch (err) {
		console.error("Failed to search entities:", err);
		return json({ error: "Failed to search entities" }, { status: 500 });
	}
};
