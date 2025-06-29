// @generated by protoc-gen-es v2.5.2 with parameter "target=ts"
// @generated from file auth/auth.proto (package auth, syntax proto3)
/* eslint-disable */

import type { Message } from "@bufbuild/protobuf";
import type {
	GenFile,
	GenMessage,
	GenService,
} from "@bufbuild/protobuf/codegenv2";
import {
	fileDesc,
	messageDesc,
	serviceDesc,
} from "@bufbuild/protobuf/codegenv2";

/**
 * Describes the file auth/auth.proto.
 */
export const file_auth_auth: GenFile =
	/*@__PURE__*/
	fileDesc(
		"Cg9hdXRoL2F1dGgucHJvdG8SBGF1dGgiMgoZUmVmcmVzaEFjY2Vzc1Rva2VuUmVxdWVzdBIVCg1yZWZyZXNoX3Rva2VuGAEgASgJIkoKGVJlZ2lzdGVyQnlQYXNzd29yZFJlcXVlc3QSDQoFZW1haWwYASABKAkSEAoIcGFzc3dvcmQYAiABKAkSDAoEbmFtZRgDIAEoCSI5ChZMb2dpbkJ5UGFzc3dvcmRSZXF1ZXN0Eg0KBWVtYWlsGAEgASgJEhAKCHBhc3N3b3JkGAIgASgJImMKDEF1dGhSZXNwb25zZRIUCgxhY2Nlc3NfdG9rZW4YASABKAkSEgoKdG9rZW5fdHlwZRgCIAEoCRISCgpleHBpcmVzX2luGAMgASgFEhUKDXJlZnJlc2hfdG9rZW4YBCABKAky6AEKC0F1dGhTZXJ2aWNlEkkKElJlZ2lzdGVyQnlQYXNzd29yZBIfLmF1dGguUmVnaXN0ZXJCeVBhc3N3b3JkUmVxdWVzdBoSLmF1dGguQXV0aFJlc3BvbnNlEkMKD0xvZ2luQnlQYXNzd29yZBIcLmF1dGguTG9naW5CeVBhc3N3b3JkUmVxdWVzdBoSLmF1dGguQXV0aFJlc3BvbnNlEkkKElJlZnJlc2hBY2Nlc3NUb2tlbhIfLmF1dGguUmVmcmVzaEFjY2Vzc1Rva2VuUmVxdWVzdBoSLmF1dGguQXV0aFJlc3BvbnNlQkBaPmdpdGh1Yi5jb20vcHJvamVjdC1taWthbi91bWkubWlrYW4vYmFja2VuZC9pbmZyYXN0cnVjdHVyZS9ncnBjYgZwcm90bzM",
	);

/**
 * アクセストークン更新用のリクエスト
 *
 * @generated from message auth.RefreshAccessTokenRequest
 */
export type RefreshAccessTokenRequest =
	Message<"auth.RefreshAccessTokenRequest"> & {
		/**
		 * @generated from field: string refresh_token = 1;
		 */
		refreshToken: string;
	};

/**
 * Describes the message auth.RefreshAccessTokenRequest.
 * Use `create(RefreshAccessTokenRequestSchema)` to create a new message.
 */
export const RefreshAccessTokenRequestSchema: GenMessage<RefreshAccessTokenRequest> =
	/*@__PURE__*/
	messageDesc(file_auth_auth, 0);

/**
 * パスワード新規登録用のリクエスト
 *
 * @generated from message auth.RegisterByPasswordRequest
 */
export type RegisterByPasswordRequest =
	Message<"auth.RegisterByPasswordRequest"> & {
		/**
		 * @generated from field: string email = 1;
		 */
		email: string;

		/**
		 * @generated from field: string password = 2;
		 */
		password: string;

		/**
		 * @generated from field: string name = 3;
		 */
		name: string;
	};

/**
 * Describes the message auth.RegisterByPasswordRequest.
 * Use `create(RegisterByPasswordRequestSchema)` to create a new message.
 */
export const RegisterByPasswordRequestSchema: GenMessage<RegisterByPasswordRequest> =
	/*@__PURE__*/
	messageDesc(file_auth_auth, 1);

/**
 * パスワードログイン用のリクエスト
 *
 * @generated from message auth.LoginByPasswordRequest
 */
export type LoginByPasswordRequest = Message<"auth.LoginByPasswordRequest"> & {
	/**
	 * @generated from field: string email = 1;
	 */
	email: string;

	/**
	 * @generated from field: string password = 2;
	 */
	password: string;
};

/**
 * Describes the message auth.LoginByPasswordRequest.
 * Use `create(LoginByPasswordRequestSchema)` to create a new message.
 */
export const LoginByPasswordRequestSchema: GenMessage<LoginByPasswordRequest> =
	/*@__PURE__*/
	messageDesc(file_auth_auth, 2);

/**
 * レスポンスはログイン方法に関わらず共通
 *
 * @generated from message auth.AuthResponse
 */
export type AuthResponse = Message<"auth.AuthResponse"> & {
	/**
	 * @generated from field: string access_token = 1;
	 */
	accessToken: string;

	/**
	 * @generated from field: string token_type = 2;
	 */
	tokenType: string;

	/**
	 * 秒単位
	 *
	 * @generated from field: int32 expires_in = 3;
	 */
	expiresIn: number;

	/**
	 * @generated from field: string refresh_token = 4;
	 */
	refreshToken: string;
};

/**
 * Describes the message auth.AuthResponse.
 * Use `create(AuthResponseSchema)` to create a new message.
 */
export const AuthResponseSchema: GenMessage<AuthResponse> =
	/*@__PURE__*/
	messageDesc(file_auth_auth, 3);

/**
 * @generated from service auth.AuthService
 */
export const AuthService: GenService<{
	/**
	 * 新規登録
	 *
	 * @generated from rpc auth.AuthService.RegisterByPassword
	 */
	registerByPassword: {
		methodKind: "unary";
		input: typeof RegisterByPasswordRequestSchema;
		output: typeof AuthResponseSchema;
	};
	/**
	 * ログイン
	 *
	 * @generated from rpc auth.AuthService.LoginByPassword
	 */
	loginByPassword: {
		methodKind: "unary";
		input: typeof LoginByPasswordRequestSchema;
		output: typeof AuthResponseSchema;
	};
	/**
	 * AccessTokenの更新
	 *
	 * @generated from rpc auth.AuthService.RefreshAccessToken
	 */
	refreshAccessToken: {
		methodKind: "unary";
		input: typeof RefreshAccessTokenRequestSchema;
		output: typeof AuthResponseSchema;
	};
}> = /*@__PURE__*/ serviceDesc(file_auth_auth, 0);
