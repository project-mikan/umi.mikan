import { writable } from "svelte/store";

export const isAuthenticated = writable(false);
export const user = writable(null);

export function initAuth() {
	// No longer needed as authentication is handled server-side
}

export function logout() {
	// This will be handled through form action in +layout.server.ts
	isAuthenticated.set(false);
	user.set(null);
}
