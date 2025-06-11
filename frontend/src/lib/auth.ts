import { LoginByPasswordRequest, RegisterByPasswordRequest, AuthResponse } from '$lib/proto/auth/auth_pb';
import { AuthServiceClient } from '$lib/proto/auth/auth_grpc_pb';
import * as grpc from '@grpc/grpc-js';
import { writable } from 'svelte/store';

export const isAuthenticated = writable(false);
export const user = writable(null);

const client = new AuthServiceClient('localhost:8080', grpc.credentials.createInsecure());

export function initAuth() {
	if (typeof window !== 'undefined') {
		const token = localStorage.getItem('accessToken');
		if (token) {
			isAuthenticated.set(true);
		}
	}
}

export function login(email: string, password: string): Promise<AuthResponse> {
	const request = new LoginByPasswordRequest();
	request.setEmail(email);
	request.setPassword(password);

	return new Promise((resolve, reject) => {
		client.loginByPassword(request, (err: grpc.ServiceError | null, response: AuthResponse) => {
			if (err) {
				reject(err);
			} else {
				localStorage.setItem('accessToken', response.getAccessToken());
				localStorage.setItem('refreshToken', response.getRefreshToken());
				isAuthenticated.set(true);
				resolve(response);
			}
		});
	});
}

export function register(name: string, email: string, password: string): Promise<AuthResponse> {
	const request = new RegisterByPasswordRequest();
	request.setName(name);
	request.setEmail(email);
	request.setPassword(password);

	return new Promise((resolve, reject) => {
		client.registerByPassword(request, (err: grpc.ServiceError | null, response: AuthResponse) => {
			if (err) {
				reject(err);
			} else {
				localStorage.setItem('accessToken', response.getAccessToken());
				localStorage.setItem('refreshToken', response.getRefreshToken());
				isAuthenticated.set(true);
				resolve(response);
			}
		});
	});
}

export function logout() {
	localStorage.removeItem('accessToken');
	localStorage.removeItem('refreshToken');
	isAuthenticated.set(false);
	user.set(null);
}