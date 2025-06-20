import { loginByPassword } from "$lib/server/auth-api";
import { fail, redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = ({ cookies }) => {
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

		if (!(email && password)) {
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
			// Log error for debugging but don't expose details to client
			return fail(400, {
				error: error instanceof Error ? error.message : "Login failed",
			});
		}

		redirect(303, "/");
	},
};
