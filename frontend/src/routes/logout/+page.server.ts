import { redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async () => {
	// Redirect to login page if accessed directly
	redirect(302, "/login");
};

export const actions: Actions = {
	default: async ({ cookies }) => {
		cookies.delete("accessToken", { path: "/" });
		cookies.delete("refreshToken", { path: "/" });
		redirect(303, "/login");
	},
};