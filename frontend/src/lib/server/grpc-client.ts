import * as grpc from '@grpc/grpc-js';
import { AuthServiceClient } from '../grpc/auth/auth_grpc_pb';
import { DiaryServiceClient } from '../grpc/diary/diary_grpc_pb';

// Backend gRPC server address
const GRPC_SERVER_ADDRESS = process.env.GRPC_SERVER_ADDRESS || 'backend:8080';

// Create gRPC clients
export function createAuthClient() {
	return new AuthServiceClient(
		GRPC_SERVER_ADDRESS,
		grpc.credentials.createInsecure()
	);
}

export function createDiaryClient() {
	return new DiaryServiceClient(
		GRPC_SERVER_ADDRESS,
		grpc.credentials.createInsecure()
	);
}

// Utility function to promisify gRPC calls
export function promisifyGrpcCall<TRequest, TResponse>(
	client: any,
	method: string,
	request: TRequest,
	metadata?: grpc.Metadata
): Promise<TResponse> {
	return new Promise((resolve, reject) => {
		const call = client[method];
		if (!call) {
			reject(new Error(`Method ${method} not found on client`));
			return;
		}

		call.call(
			client,
			request,
			metadata || new grpc.Metadata(),
			(error: grpc.ServiceError | null, response: TResponse) => {
				if (error) {
					reject(error);
				} else {
					resolve(response);
				}
			}
		);
	});
}