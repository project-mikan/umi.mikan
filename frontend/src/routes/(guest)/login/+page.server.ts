import { fail, redirect } from "@sveltejs/kit";
import { loginByPassword } from "$lib/server/auth-api";
import { translateErrorMessage, isRateLimitError } from "$lib/utils/error-utils";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ parent }) => {
	const { isAuthenticated } = await parent();
	if (isAuthenticated) {
		throw redirect(302, "/");
	}
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const email = data.get("email") as string;
		const password = data.get("password") as string;

		if (!email || !password) {
			return fail(400, { error: "Email and password are required" });
		}

		try {
			const response = await loginByPassword({
				email,
				password,
			});

			cookies.set("accessToken", response.accessToken, {
				path: "/",
				httpOnly: true,
				secure: false, // Set to true in production with HTTPS
				sameSite: "strict",
				maxAge: 60 * 15, // 15 minutes
			});

			cookies.set("refreshToken", response.refreshToken, {
				path: "/",
				httpOnly: true,
				secure: false, // Set to true in production with HTTPS
				sameSite: "strict",
				maxAge: 60 * 60 * 24 * 30, // 30 days
			});
		} catch (error: unknown) {
			console.error("Login error:", error);

			// レート制限エラーの場合は429ステータスを返す
			const statusCode = isRateLimitError(error) ? 429 : 400;

			return fail(statusCode, {
				error: translateErrorMessage(error),
				isRateLimited: isRateLimitError(error),
			});
		}

		redirect(303, "/");
	},
};
