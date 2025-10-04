import { error, fail, redirect } from "@sveltejs/kit";
import {
	getEntity,
	updateEntity,
	deleteEntity,
	createEntityAlias,
	deleteEntityAlias,
	getDiariesByEntity,
} from "$lib/server/entity-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { EntityCategory } from "$lib/grpc/entity/entity_pb";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ params, cookies }) => {
	if (!params.id || params.id.trim() === "") {
		throw error(500, "Invalid entity ID");
	}

	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw redirect(302, "/login");
	}

	try {
		const entityResponse = await getEntity({
			id: params.id,
			accessToken: authResult.accessToken,
		});

		const diariesResponse = await getDiariesByEntity({
			entityId: params.id,
			accessToken: authResult.accessToken,
		});

		return {
			entity: entityResponse.entity,
			diaries: diariesResponse.diaries,
		};
	} catch (err) {
		console.error("Failed to load entity:", err);
		throw redirect(302, "/entities");
	}
};

export const actions: Actions = {
	// エンティティ更新
	updateEntity: async ({ request, params, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);
		if (!authResult.isAuthenticated || !authResult.accessToken) {
			return fail(401, { error: "unauthorized", action: "updateEntity" });
		}

		const data = await request.formData();
		const name = data.get("name")?.toString();
		const categoryStr = data.get("category")?.toString();
		const memo = data.get("memo")?.toString() || "";

		if (!name) {
			return fail(400, {
				error: "nameRequired",
				action: "updateEntity",
			});
		}

		const category =
			categoryStr === "1" ? EntityCategory.PEOPLE : EntityCategory.NO_CATEGORY;

		try {
			await updateEntity({
				id: params.id,
				name,
				category,
				memo,
				accessToken: authResult.accessToken,
			});

			return {
				success: true,
				message: "updateSuccess",
				action: "updateEntity",
			};
		} catch (err) {
			console.error("Failed to update entity:", err);
			return fail(500, {
				error: "error",
				action: "updateEntity",
			});
		}
	},

	// エンティティ削除
	deleteEntity: async ({ params, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);
		if (!authResult.isAuthenticated || !authResult.accessToken) {
			return fail(401, { error: "unauthorized", action: "deleteEntity" });
		}

		try {
			await deleteEntity({
				id: params.id,
				accessToken: authResult.accessToken,
			});

			throw redirect(302, "/entities");
		} catch (err) {
			if (err instanceof Response) {
				throw err;
			}
			console.error("Failed to delete entity:", err);
			return fail(500, {
				error: "error",
				action: "deleteEntity",
			});
		}
	},

	// エイリアス追加
	createAlias: async ({ request, params, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);
		if (!authResult.isAuthenticated || !authResult.accessToken) {
			return fail(401, { error: "unauthorized", action: "createAlias" });
		}

		const data = await request.formData();
		const alias = data.get("alias")?.toString();

		if (!alias) {
			return fail(400, {
				error: "nameRequired",
				action: "createAlias",
			});
		}

		try {
			await createEntityAlias({
				entityId: params.id,
				alias,
				accessToken: authResult.accessToken,
			});

			return {
				success: true,
				message: "aliasCreateSuccess",
				action: "createAlias",
			};
		} catch (err) {
			console.error("Failed to create alias:", err);

			// エラーメッセージを判定
			if (err instanceof Error) {
				// エンティティ名として使用されているエラー
				if (err.message.includes("already used as an entity name")) {
					return fail(400, {
						error: "aliasUsedAsName",
						action: "createAlias",
					});
				}

				// エイリアスとして既に使用されているエラー
				if (
					err.message.includes("already used") ||
					err.message.includes("already exists")
				) {
					return fail(400, {
						error: "aliasAlreadyExists",
						action: "createAlias",
					});
				}
			}

			return fail(500, {
				error: "error",
				action: "createAlias",
			});
		}
	},

	// エイリアス削除
	deleteAlias: async ({ request, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);
		if (!authResult.isAuthenticated || !authResult.accessToken) {
			return fail(401, { error: "unauthorized", action: "deleteAlias" });
		}

		const data = await request.formData();
		const aliasId = data.get("aliasId")?.toString();

		if (!aliasId) {
			return fail(400, {
				error: "error",
				action: "deleteAlias",
			});
		}

		try {
			await deleteEntityAlias({
				id: aliasId,
				accessToken: authResult.accessToken,
			});

			return {
				success: true,
				message: "aliasDeleteSuccess",
				action: "deleteAlias",
			};
		} catch (err) {
			console.error("Failed to delete alias:", err);
			return fail(500, {
				error: "error",
				action: "deleteAlias",
			});
		}
	},
};
