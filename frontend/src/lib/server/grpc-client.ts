// Simplified gRPC client using direct HTTP calls
// This is a minimal implementation for the protobuf-es generated code

// Backend gRPC server address  
const GRPC_SERVER_ADDRESS = process.env.GRPC_SERVER_ADDRESS || 'http://backend:8080';

// Simple HTTP-based gRPC client
async function callGrpcEndpoint(path: string, data: any): Promise<any> {
	const response = await fetch(`${GRPC_SERVER_ADDRESS}${path}`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			'Accept': 'application/json'
		},
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		throw new Error(`gRPC call failed: ${response.status} ${response.statusText}`);
	}

	return await response.json();
}

// Auth service methods
export const authService = {
	registerByPassword: async (input: { email: string; password: string; name: string }) => {
		return callGrpcEndpoint('/auth.AuthService/RegisterByPassword', input);
	},
	loginByPassword: async (input: { email: string; password: string }) => {
		return callGrpcEndpoint('/auth.AuthService/LoginByPassword', input);
	},
	refreshAccessToken: async (input: { refreshToken: string }) => {
		return callGrpcEndpoint('/auth.AuthService/RefreshAccessToken', input);
	}
};

// Diary service methods
export const diaryService = {
	createDiaryEntry: async (input: { content: string; date: any }) => {
		return callGrpcEndpoint('/diary.DiaryService/CreateDiaryEntry', input);
	},
	getDiaryEntry: async (input: { date: any }) => {
		return callGrpcEndpoint('/diary.DiaryService/GetDiaryEntry', input);
	}
};