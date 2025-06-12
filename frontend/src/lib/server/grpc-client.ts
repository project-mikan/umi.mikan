import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import type { ProtoGrpcType as AuthProtoGrpcType } from '../grpc/auth';
import type { ProtoGrpcType as DiaryProtoGrpcType } from '../grpc/diary';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Backend gRPC server address
const GRPC_SERVER_ADDRESS = 'localhost:8080';

// Proto file paths (relative to project root)
const AUTH_PROTO_PATH = join(__dirname, '../../../../../proto/auth/auth.proto');
const DIARY_PROTO_PATH = join(__dirname, '../../../../../proto/diary/diary.proto');

// Package definition options
const packageDefinition = protoLoader.loadSync([AUTH_PROTO_PATH, DIARY_PROTO_PATH], {
	keepCase: true,
	longs: String,
	enums: String,
	defaults: true,
	oneofs: true,
});

// Load the proto definitions
const authProto = grpc.loadPackageDefinition(packageDefinition) as unknown as AuthProtoGrpcType;
const diaryProto = grpc.loadPackageDefinition(packageDefinition) as unknown as DiaryProtoGrpcType;

// Create gRPC clients
export function createAuthClient() {
	return new authProto.auth.AuthService(
		GRPC_SERVER_ADDRESS,
		grpc.credentials.createInsecure()
	);
}

export function createDiaryClient() {
	return new diaryProto.diary.DiaryService(
		GRPC_SERVER_ADDRESS,
		grpc.credentials.createInsecure()
	);
}

// Utility function to promisify gRPC calls
export function promisifyGrpcCall<TRequest, TResponse>(
	client: grpc.Client,
	method: string,
	request: TRequest,
	metadata?: grpc.Metadata
): Promise<TResponse> {
	return new Promise((resolve, reject) => {
		const call = (client as any)[method];
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