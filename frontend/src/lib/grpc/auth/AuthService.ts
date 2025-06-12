// Original file: proto/auth/auth.proto

import type * as grpc from "@grpc/grpc-js";
import type { MethodDefinition } from "@grpc/proto-loader";
import type {
	AuthResponse as _auth_AuthResponse,
	AuthResponse__Output as _auth_AuthResponse__Output,
} from "../auth/AuthResponse";
import type {
	LoginByPasswordRequest as _auth_LoginByPasswordRequest,
	LoginByPasswordRequest__Output as _auth_LoginByPasswordRequest__Output,
} from "../auth/LoginByPasswordRequest";
import type {
	RefreshAccessTokenRequest as _auth_RefreshAccessTokenRequest,
	RefreshAccessTokenRequest__Output as _auth_RefreshAccessTokenRequest__Output,
} from "../auth/RefreshAccessTokenRequest";
import type {
	RegisterByPasswordRequest as _auth_RegisterByPasswordRequest,
	RegisterByPasswordRequest__Output as _auth_RegisterByPasswordRequest__Output,
} from "../auth/RegisterByPasswordRequest";

export interface AuthServiceClient extends grpc.Client {
	LoginByPassword(
		argument: _auth_LoginByPasswordRequest,
		metadata: grpc.Metadata,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	LoginByPassword(
		argument: _auth_LoginByPasswordRequest,
		metadata: grpc.Metadata,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	LoginByPassword(
		argument: _auth_LoginByPasswordRequest,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	LoginByPassword(
		argument: _auth_LoginByPasswordRequest,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	loginByPassword(
		argument: _auth_LoginByPasswordRequest,
		metadata: grpc.Metadata,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	loginByPassword(
		argument: _auth_LoginByPasswordRequest,
		metadata: grpc.Metadata,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	loginByPassword(
		argument: _auth_LoginByPasswordRequest,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	loginByPassword(
		argument: _auth_LoginByPasswordRequest,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;

	RefreshAccessToken(
		argument: _auth_RefreshAccessTokenRequest,
		metadata: grpc.Metadata,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	RefreshAccessToken(
		argument: _auth_RefreshAccessTokenRequest,
		metadata: grpc.Metadata,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	RefreshAccessToken(
		argument: _auth_RefreshAccessTokenRequest,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	RefreshAccessToken(
		argument: _auth_RefreshAccessTokenRequest,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	refreshAccessToken(
		argument: _auth_RefreshAccessTokenRequest,
		metadata: grpc.Metadata,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	refreshAccessToken(
		argument: _auth_RefreshAccessTokenRequest,
		metadata: grpc.Metadata,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	refreshAccessToken(
		argument: _auth_RefreshAccessTokenRequest,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	refreshAccessToken(
		argument: _auth_RefreshAccessTokenRequest,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;

	RegisterByPassword(
		argument: _auth_RegisterByPasswordRequest,
		metadata: grpc.Metadata,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	RegisterByPassword(
		argument: _auth_RegisterByPasswordRequest,
		metadata: grpc.Metadata,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	RegisterByPassword(
		argument: _auth_RegisterByPasswordRequest,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	RegisterByPassword(
		argument: _auth_RegisterByPasswordRequest,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	registerByPassword(
		argument: _auth_RegisterByPasswordRequest,
		metadata: grpc.Metadata,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	registerByPassword(
		argument: _auth_RegisterByPasswordRequest,
		metadata: grpc.Metadata,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	registerByPassword(
		argument: _auth_RegisterByPasswordRequest,
		options: grpc.CallOptions,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
	registerByPassword(
		argument: _auth_RegisterByPasswordRequest,
		callback: grpc.requestCallback<_auth_AuthResponse__Output>,
	): grpc.ClientUnaryCall;
}

export interface AuthServiceHandlers extends grpc.UntypedServiceImplementation {
	LoginByPassword: grpc.handleUnaryCall<
		_auth_LoginByPasswordRequest__Output,
		_auth_AuthResponse
	>;

	RefreshAccessToken: grpc.handleUnaryCall<
		_auth_RefreshAccessTokenRequest__Output,
		_auth_AuthResponse
	>;

	RegisterByPassword: grpc.handleUnaryCall<
		_auth_RegisterByPasswordRequest__Output,
		_auth_AuthResponse
	>;
}

export interface AuthServiceDefinition extends grpc.ServiceDefinition {
	LoginByPassword: MethodDefinition<
		_auth_LoginByPasswordRequest,
		_auth_AuthResponse,
		_auth_LoginByPasswordRequest__Output,
		_auth_AuthResponse__Output
	>;
	RefreshAccessToken: MethodDefinition<
		_auth_RefreshAccessTokenRequest,
		_auth_AuthResponse,
		_auth_RefreshAccessTokenRequest__Output,
		_auth_AuthResponse__Output
	>;
	RegisterByPassword: MethodDefinition<
		_auth_RegisterByPasswordRequest,
		_auth_AuthResponse,
		_auth_RegisterByPasswordRequest__Output,
		_auth_AuthResponse__Output
	>;
}
