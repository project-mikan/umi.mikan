// Original file: proto/auth/auth.proto

export interface RegisterByPasswordRequest {
	email?: string;
	password?: string;
	name?: string;
}

export interface RegisterByPasswordRequest__Output {
	email: string;
	password: string;
	name: string;
}
