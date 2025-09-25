/**
 * Convert Unix timestamp (seconds) to JavaScript timestamp (milliseconds)
 */
export function unixToMilliseconds(
	unixTimestamp: number | string | bigint,
): number {
	return Number(unixTimestamp) * 1000;
}

/**
 * Convert minutes to milliseconds
 */
export function minutesToMilliseconds(minutes: number): number {
	return minutes * 60 * 1000;
}

/**
 * Get total milliseconds in a day
 */
export function getDayInMilliseconds(): number {
	return 24 * 60 * 60 * 1000;
}

export function isTokenExpiringSoon(token: string, bufferMinutes = 5): boolean {
	try {
		const payload = JSON.parse(atob(token.split(".")[1]));
		const expiryTime = unixToMilliseconds(payload.exp);
		const now = Date.now();
		const bufferTime = minutesToMilliseconds(bufferMinutes);

		return expiryTime - now < bufferTime;
	} catch {
		return true;
	}
}

/**
 * Parse JWT token and return payload
 */
export function parseJWT(token: string): { sub?: string; exp?: number } | null {
	try {
		const payload = JSON.parse(atob(token.split(".")[1]));
		return payload;
	} catch {
		return null;
	}
}
