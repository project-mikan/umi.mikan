export function isTokenExpiringSoon(token: string, bufferMinutes = 5): boolean {
	try {
		const payload = JSON.parse(atob(token.split(".")[1]));
		// Convert JWT exp claim (Unix timestamp in seconds) to JavaScript timestamp (milliseconds)
		const expiryTime = payload.exp * 1000;
		const now = Date.now();
		// Convert buffer minutes to milliseconds
		const bufferTime = bufferMinutes * 60 * 1000;

		return expiryTime - now < bufferTime;
	} catch {
		return true;
	}
}
