import { loginByPassword } from "$lib/server/auth-api.js";
import { fail, redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies }) => {
	const accessToken = cookies.get("accessToken");
	if (accessToken) {
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

			cookies.set("accessToken", response.access_token, {
				path: "/",
				httpOnly: true,
				secure: false, // Set to true in production with HTTPS
				sameSite: "strict",
				maxAge: 60 * 15, // 15 minutes
			});

			cookies.set("refreshToken", response.refresh_token, {
				path: "/",
				httpOnly: true,
				secure: false, // Set to true in production with HTTPS
				sameSite: "strict",
				maxAge: 60 * 60 * 24 * 30, // 30 days
			});

			throw redirect(302, "/");
		} catch (error: unknown) {
			console.error("Login error:", error);
			return fail(400, { error: error instanceof Error ? error.message : "Login failed" });
		}
	},
};
