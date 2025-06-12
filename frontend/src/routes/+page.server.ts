import { redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies }) => {
	const accessToken = cookies.get("accessToken");
	const isAuthenticated = !!accessToken;

	if (!isAuthenticated) {
		throw redirect(302, "/login");
	}

	return {
		isAuthenticated,
	};
};

export const actions: Actions = {
	logout: async ({ cookies }) => {
		cookies.delete("accessToken", { path: "/" });
		cookies.delete("refreshToken", { path: "/" });
		throw redirect(302, "/login");
	},
};
