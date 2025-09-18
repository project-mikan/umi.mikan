import { redirect, fail } from "@sveltejs/kit";
import {
	updateUserName,
	changePassword,
	updateLLMKey,
	getUserInfo,
	deleteLLMKey,
	deleteAccount,
} from "$lib/server/auth-api";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw redirect(302, "/login");
	}

	try {
		const userInfo = await getUserInfo({ accessToken });
		return {
			user: {
				name: userInfo.name,
				email: userInfo.email,
				llmKeys: userInfo.llmKeys || [],
			},
		};
	} catch (error) {
		console.error("Failed to get user info:", error);
		throw redirect(302, "/login");
	}
};

export const actions: Actions = {
	updateUsername: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "unauthorized" });
		}

		const data = await request.formData();
		const newName = data.get("username") as string;

		if (!newName || newName.trim() === "") {
			return fail(400, { error: "nameRequired" });
		}

		if (newName.length > 20) {
			return fail(400, { error: "nameTooLong" });
		}

		try {
			const response = await updateUserName({
				newName: newName.trim(),
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message });
			}

			return { success: true, message: response.message };
		} catch (error) {
			console.error("Update username error:", error);
			return fail(500, { error: "updateFailed" });
		}
	},

	changePassword: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "unauthorized" });
		}

		const data = await request.formData();
		const currentPassword = data.get("currentPassword") as string;
		const newPassword = data.get("newPassword") as string;
		const confirmPassword = data.get("confirmPassword") as string;

		if (!currentPassword || !newPassword || !confirmPassword) {
			return fail(400, { error: "passwordsRequired" });
		}

		if (newPassword !== confirmPassword) {
			return fail(400, { error: "passwordsDoNotMatch" });
		}

		if (newPassword.length < 8) {
			return fail(400, { error: "passwordTooShort" });
		}

		try {
			const response = await changePassword({
				currentPassword,
				newPassword,
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message });
			}

			return { success: true, message: response.message };
		} catch (error) {
			console.error("Change password error:", error);
			return fail(500, { error: "updateFailed" });
		}
	},

	updateLLMKey: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "unauthorized" });
		}

		const data = await request.formData();
		const llmProvider = parseInt(data.get("llmProvider") as string, 10);
		const key = data.get("llmKey") as string;

		if (Number.isNaN(llmProvider) || llmProvider < 0) {
			return fail(400, { error: "invalidProvider" });
		}

		if (!key || key.trim() === "") {
			return fail(400, { error: "tokenRequired" });
		}

		if (key.length > 100) {
			return fail(400, { error: "tokenTooLong" });
		}

		try {
			const response = await updateLLMKey({
				llmProvider,
				key: key.trim(),
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message });
			}

			return { success: true, message: response.message };
		} catch (error) {
			console.error("Update LLM token error:", error);
			return fail(500, { error: "updateFailed" });
		}
	},

	deleteLLMKey: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "unauthorized" });
		}

		const data = await request.formData();
		const llmProvider = parseInt(data.get("llmProvider") as string, 10);

		if (Number.isNaN(llmProvider) || llmProvider < 0) {
			return fail(400, { error: "invalidProvider" });
		}

		try {
			const response = await deleteLLMKey({
				llmProvider,
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message });
			}

			return { success: true, message: response.message };
		} catch (error) {
			console.error("Delete LLM token error:", error);
			return fail(500, { error: "updateFailed" });
		}
	},

	deleteAccount: async ({ cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "unauthorized" });
		}

		try {
			const response = await deleteAccount({
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message });
			}

			// Clear cookies after successful account deletion
			cookies.set("accessToken", "", {
				path: "/",
				expires: new Date(0),
				httpOnly: true,
				secure: true,
				sameSite: "strict",
			});

			cookies.set("refreshToken", "", {
				path: "/",
				expires: new Date(0),
				httpOnly: true,
				secure: true,
				sameSite: "strict",
			});

			// Redirect to login page after account deletion
			throw redirect(302, "/login");
		} catch (error) {
			if (
				error &&
				typeof error === "object" &&
				"status" in error &&
				error.status === 302
			) {
				// Re-throw redirect
				throw error;
			}
			console.error("Delete account error:", error);
			return fail(500, { error: "updateFailed" });
		}
	},
};
