import { browser } from "$app/environment";
import { goto } from "$app/navigation";

/**
 * Client-side authentication utilities for token management
 */

interface RefreshTokenResponse {
	accessToken: string;
	refreshToken?: string;
}

/**
 * Attempts to refresh the access token by calling the refresh API endpoint
 */
export async function refreshAccessToken(): Promise<string | null> {
	if (!browser) return null;

	try {
		const response = await fetch("/api/auth/refresh", {
			method: "POST",
			credentials: "include",
		});

		if (response.ok) {
			const data: RefreshTokenResponse = await response.json();
			return data.accessToken;
		}

		if (response.status === 401) {
			// Refresh token is also expired, redirect to login
			await goto("/login");
			return null;
		}

		console.error("Token refresh failed with status:", response.status);
		return null;
	} catch (error) {
		console.error("Token refresh error:", error);
		return null;
	}
}

/**
 * Enhanced fetch function that automatically handles token refresh on 401 errors
 */
export async function authenticatedFetch(
	input: RequestInfo | URL,
	init?: RequestInit,
): Promise<Response> {
	if (!browser) {
		throw new Error(
			"authenticatedFetch can only be used in browser environment",
		);
	}

	// First attempt
	let response = await fetch(input, {
		...init,
		credentials: "include",
	});

	// If we get 401, try to refresh token and retry once
	if (response.status === 401) {
		console.log("Received 401, attempting token refresh...");
		const newAccessToken = await refreshAccessToken();

		if (newAccessToken) {
			console.log("Token refresh successful, retrying original request...");
			// Retry the original request
			response = await fetch(input, {
				...init,
				credentials: "include",
			});
		} else {
			console.log("Token refresh failed, returning 401 response");
		}
	}

	return response;
}
