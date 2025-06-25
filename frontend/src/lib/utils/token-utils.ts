export function isTokenExpiringSoon(token: string, bufferMinutes = 5): boolean {
	try {
		const payload = JSON.parse(atob(token.split(".")[1]));
		const expiryTime = payload.exp * 1000;
		const now = Date.now();
		const bufferTime = bufferMinutes * 60 * 1000;

		return expiryTime - now < bufferTime;
	} catch {
		return true;
	}
}
