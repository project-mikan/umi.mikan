import { redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = ({ cookies }) => {
	const accessToken = cookies.get("accessToken");
	const isAuthenticated = !!accessToken;

	return {
		isAuthenticated,
	};
};

export const actions: Actions = {
	logout: ({ cookies }) => {
		cookies.delete("accessToken", { path: "/" });
		cookies.delete("refreshToken", { path: "/" });
		throw redirect(302, "/login");
	},
};
