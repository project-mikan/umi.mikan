import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = async ({ cookies, url }) => {
	const accessToken = cookies.get("accessToken");
	const isAuthenticated = !!accessToken;

	return {
		isAuthenticated,
		path: url.pathname,
	};
};
