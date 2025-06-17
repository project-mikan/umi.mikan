import { registerByPassword } from "$lib/server/auth-api";
import { fail, redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies }) => {
	const accessToken = cookies.get("accessToken");
	if (accessToken) {
		throw redirect(302, "/diary");
	}
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const name = data.get("name") as string;
		const email = data.get("email") as string;
		const password = data.get("password") as string;

		if (!name || !email || !password) {
			return fail(400, { error: "Name, email and password are required" });
		}

		try {
			const response = await registerByPassword({
				name,
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
