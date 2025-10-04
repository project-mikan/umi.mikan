import type { Actions, PageServerLoad } from "./$types";
import { fail, redirect } from "@sveltejs/kit";
import { createEntity } from "$lib/server/entity-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { EntityCategory } from "$lib/grpc/entity/entity_pb";

export const load: PageServerLoad = async ({ cookies }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw redirect(302, "/login");
	}

	return {};
};

export const actions = {
	create: async ({ request, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);

		if (!authResult.isAuthenticated || !authResult.accessToken) {
			throw redirect(302, "/login");
		}

		const formData = await request.formData();
		const name = formData.get("name") as string;
		const categoryStr = formData.get("category") as string;
		const memo = formData.get("memo") as string;

		// バリデーション
		if (!name || name.trim() === "") {
			return fail(400, {
				error: "entity.messages.nameRequired",
				name,
				category: categoryStr,
				memo,
			});
		}

		// カテゴリを変換
		const category =
			categoryStr === "people" || categoryStr === "1"
				? EntityCategory.PEOPLE
				: EntityCategory.NO_CATEGORY;

		// エンティティ作成
		const response = await createEntity({
			name: name.trim(),
			category,
			memo: memo?.trim() || "",
			accessToken: authResult.accessToken,
		}).catch((error) => {
			// gRPCエラーのハンドリング
			if (error instanceof Error) {
				// エラーメッセージをコンソールに出力
				console.error("Entity creation error:", error.message);

				// AlreadyExistsエラー（重複）をチェック
				if (error.message.includes("already exists")) {
					return fail(400, {
						error: "entity.messages.duplicateName",
						name,
						category: categoryStr,
						memo,
					});
				}

				// エイリアスとして使用されているエラー
				if (error.message.includes("already used as an alias")) {
					return fail(400, {
						error: "entity.messages.nameUsedAsAlias",
						name,
						category: categoryStr,
						memo,
					});
				}

				// エラーの詳細情報を取得
				let errorDetail = error.message;
				// ConnectErrorの場合、追加情報を含める
				if ("code" in error && "rawMessage" in error) {
					const connectError = error as Error & {
						code: string;
						rawMessage?: string;
					};
					errorDetail = `[${connectError.code}] ${connectError.rawMessage || error.message}`;
				}

				// 詳細なエラーメッセージを返す
				return fail(500, {
					error: "entity.messages.error",
					errorDetail,
					name,
					category: categoryStr,
					memo,
				});
			}

			// その他のエラー
			console.error("Unknown error:", error);
			return fail(500, {
				error: "entity.messages.error",
				errorDetail: typeof error === "string" ? error : JSON.stringify(error),
				name,
				category: categoryStr,
				memo,
			});
		});

		// エラーレスポンスの場合はそのまま返す
		if (response && "status" in response) {
			return response;
		}

		// 作成したエンティティの詳細ページへリダイレクト
		if (!response.entity?.id) {
			return fail(500, {
				error: "entity.messages.error",
				name,
				category: categoryStr,
				memo,
			});
		}

		redirect(303, `/entity/${response.entity.id}`);
	},
} satisfies Actions;
