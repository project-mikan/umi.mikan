import { fail, redirect } from "@sveltejs/kit";
import { loginByPassword } from "$lib/server/auth-api";
import {
	translateErrorMessage,
	isRateLimitError,
} from "$lib/utils/error-utils";
import {
	ACCESS_TOKEN_COOKIE_OPTIONS,
	REFRESH_TOKEN_COOKIE_OPTIONS,
} from "$lib/utils/cookie-utils";
import { validateCSRFToken } from "$lib/server/csrf";
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
		const csrfToken = data.get("csrfToken") as string;

		// CSRF トークンの検証
		if (!validateCSRFToken(cookies, csrfToken)) {
			return fail(403, { error: "Invalid CSRF token" });
		}

		if (!email || !password) {
			return fail(400, { error: "Email and password are required" });
		}

		try {
			const response = await loginByPassword({
				email,
				password,
			});

			cookies.set(
				"accessToken",
				response.accessToken,
				ACCESS_TOKEN_COOKIE_OPTIONS,
			);

			cookies.set(
				"refreshToken",
				response.refreshToken,
				REFRESH_TOKEN_COOKIE_OPTIONS,
			);
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
