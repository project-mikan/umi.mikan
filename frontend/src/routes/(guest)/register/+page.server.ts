import { fail, redirect } from "@sveltejs/kit";
import { registerByPassword } from "$lib/server/auth-api";
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
		const name = data.get("name") as string;
		const email = data.get("email") as string;
		const password = data.get("password") as string;
		const csrfToken = data.get("csrfToken") as string;

		// CSRF トークンの検証
		if (!validateCSRFToken(cookies, csrfToken)) {
			return fail(403, { error: "Invalid CSRF token" });
		}

		if (!name || !email || !password) {
			return fail(400, { error: "Name, email and password are required" });
		}

		try {
			const response = await registerByPassword({
				name,
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
			console.error("Register error:", error);
			let errorMessage = "Registration failed";

			if (error instanceof Error) {
				if (error.message.includes("already exists")) {
					errorMessage = "このメールアドレスは既に登録済みです";
				} else if (error.message.includes("validation error")) {
					errorMessage = "入力内容に問題があります";
				} else {
					errorMessage = error.message;
				}
			}

			return fail(400, {
				error: errorMessage,
			});
		}

		redirect(303, "/");
	},
};
