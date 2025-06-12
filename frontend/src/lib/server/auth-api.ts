// HTTP APIを使用したバックエンドとの通信
// gRPCの代わりにHTTPで通信する簡単な実装

interface LoginRequest {
	email: string;
	password: string;
}

interface RegisterRequest {
	name: string;
	email: string;
	password: string;
}

interface AuthResponse {
	access_token: string;
	refresh_token: string;
	token_type: string;
	expires_in: number;
}

export async function loginByPassword(request: LoginRequest): Promise<AuthResponse> {
	// TODO: 実際にはバックエンドのHTTP APIエンドポイントを呼び出す
	// 現在はダミーレスポンスを返す
	await new Promise(resolve => setTimeout(resolve, 500)); // API呼び出しを模擬
	
	if (request.email === 'test@example.com' && request.password === 'password') {
		return {
			access_token: 'dummy_access_token',
			refresh_token: 'dummy_refresh_token',
			token_type: 'Bearer',
			expires_in: 900
		};
	}
	
	throw new Error('Invalid credentials');
}

export async function registerByPassword(request: RegisterRequest): Promise<AuthResponse> {
	// TODO: 実際にはバックエンドのHTTP APIエンドポイントを呼び出す
	// 現在はダミーレスポンスを返す
	await new Promise(resolve => setTimeout(resolve, 500)); // API呼び出しを模擬
	
	if (request.email && request.password && request.name) {
		return {
			access_token: 'dummy_access_token',
			refresh_token: 'dummy_refresh_token',
			token_type: 'Bearer',
			expires_in: 900
		};
	}
	
	throw new Error('Registration failed');
}