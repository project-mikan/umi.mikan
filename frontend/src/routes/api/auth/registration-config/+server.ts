import { json } from "@sveltejs/kit";
import { getRegistrationConfig } from "$lib/server/auth-api";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async () => {
	try {
		const config = await getRegistrationConfig();
		return json({
			registerKeyRequired: config.registerKeyRequired,
		});
	} catch (error) {
		console.error("Failed to get registration config:", error);
		// エラーの場合はデフォルトでfalseを返す
		return json({
			registerKeyRequired: false,
		});
	}
};
