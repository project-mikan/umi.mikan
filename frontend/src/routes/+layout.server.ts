import type { LayoutServerLoad } from "./$types.ts";

export const load: LayoutServerLoad = ({ cookies, url }) => {
	const accessToken = cookies.get("accessToken");
	const isAuthenticated = !!accessToken;

	return {
		isAuthenticated,
		path: url.pathname,
	};
};
