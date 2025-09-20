import { redirect, fail } from "@sveltejs/kit";
import {
	updateUserName,
	changePassword,
	updateLLMKey,
	getUserInfo,
	deleteLLMKey,
	deleteAccount,
	updateAutoSummarySettings,
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
			return fail(400, { error: "nameRequired", action: "updateUsername" });
		}

		if (newName.length > 20) {
			return fail(400, { error: "nameTooLong", action: "updateUsername" });
		}

		try {
			const response = await updateUserName({
				newName: newName.trim(),
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message, action: "updateUsername" });
			}

			return {
				success: true,
				message: response.message,
				action: "updateUsername",
			};
		} catch (error) {
			console.error("Update username error:", error);
			return fail(500, { error: "updateFailed", action: "updateUsername" });
		}
	},

	changePassword: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "unauthorized", action: "changePassword" });
		}

		const data = await request.formData();
		const currentPassword = data.get("currentPassword") as string;
		const newPassword = data.get("newPassword") as string;
		const confirmPassword = data.get("confirmPassword") as string;

		if (!currentPassword || !newPassword || !confirmPassword) {
			return fail(400, {
				error: "passwordsRequired",
				action: "changePassword",
			});
		}

		if (newPassword !== confirmPassword) {
			return fail(400, {
				error: "passwordsDoNotMatch",
				action: "changePassword",
			});
		}

		if (newPassword.length < 8) {
			return fail(400, { error: "passwordTooShort", action: "changePassword" });
		}

		try {
			const response = await changePassword({
				currentPassword,
				newPassword,
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message, action: "changePassword" });
			}

			return {
				success: true,
				message: response.message,
				action: "changePassword",
			};
		} catch (error) {
			console.error("Change password error:", error);
			return fail(500, { error: "updateFailed", action: "changePassword" });
		}
	},

	updateLLMKey: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "unauthorized", action: "updateLLMKey" });
		}

		const data = await request.formData();
		const llmProvider = parseInt(data.get("llmProvider") as string, 10);
		const key = data.get("llmKey") as string;

		if (Number.isNaN(llmProvider) || llmProvider < 0) {
			return fail(400, { error: "invalidProvider", action: "updateLLMKey" });
		}

		if (!key || key.trim() === "") {
			return fail(400, { error: "tokenRequired", action: "updateLLMKey" });
		}

		if (key.length > 100) {
			return fail(400, { error: "tokenTooLong", action: "updateLLMKey" });
		}

		try {
			const response = await updateLLMKey({
				llmProvider,
				key: key.trim(),
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message, action: "updateLLMKey" });
			}

			return {
				success: true,
				message: response.message,
				action: "updateLLMKey",
			};
		} catch (error) {
			console.error("Update LLM token error:", error);
			return fail(500, { error: "updateFailed", action: "updateLLMKey" });
		}
	},

	deleteLLMKey: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "unauthorized", action: "deleteLLMKey" });
		}

		const data = await request.formData();
		const llmProvider = parseInt(data.get("llmProvider") as string, 10);

		if (Number.isNaN(llmProvider) || llmProvider < 0) {
			return fail(400, { error: "invalidProvider", action: "deleteLLMKey" });
		}

		try {
			const response = await deleteLLMKey({
				llmProvider,
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message, action: "deleteLLMKey" });
			}

			return {
				success: true,
				message: response.message,
				action: "deleteLLMKey",
			};
		} catch (error) {
			console.error("Delete LLM token error:", error);
			return fail(500, { error: "updateFailed", action: "deleteLLMKey" });
		}
	},

	deleteAccount: async ({ cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "unauthorized", action: "deleteAccount" });
		}

		try {
			const response = await deleteAccount({
				accessToken,
			});

			if (!response.success) {
				return fail(400, { error: response.message, action: "deleteAccount" });
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
			return fail(500, { error: "updateFailed", action: "deleteAccount" });
		}
	},

	updateAutoSummarySettings: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, {
				error: "unauthorized",
				action: "updateAutoSummarySettings",
			});
		}

		const data = await request.formData();
		const llmProvider = parseInt(data.get("llmProvider") as string, 10);
		const autoSummaryDaily = data.get("autoSummaryDaily") === "on";
		const autoSummaryMonthly = data.get("autoSummaryMonthly") === "on";

		if (Number.isNaN(llmProvider) || llmProvider < 0) {
			return fail(400, {
				error: "invalidProvider",
				action: "updateAutoSummarySettings",
			});
		}

		try {
			const response = await updateAutoSummarySettings({
				llmProvider,
				autoSummaryDaily,
				autoSummaryMonthly,
				accessToken,
			});

			if (!response.success) {
				return fail(400, {
					error: response.message,
					action: "updateAutoSummarySettings",
				});
			}

			// Get updated user info to refresh the form state
			const userInfo = await getUserInfo({ accessToken });

			return {
				success: true,
				message: response.message,
				action: "updateAutoSummarySettings",
				user: {
					name: userInfo.name,
					email: userInfo.email,
					llmKeys: userInfo.llmKeys || [],
				},
			};
		} catch (error) {
			console.error("Update auto summary settings error:", error);
			return fail(500, {
				error: "updateFailed",
				action: "updateAutoSummarySettings",
			});
		}
	},
};
