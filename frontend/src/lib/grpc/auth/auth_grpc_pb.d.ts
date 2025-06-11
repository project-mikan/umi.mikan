// package: auth
// file: auth/auth.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as auth_auth_pb from "../auth/auth_pb";

interface IAuthServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    registerByPassword: IAuthServiceService_IRegisterByPassword;
    loginByPassword: IAuthServiceService_ILoginByPassword;
    refreshAccessToken: IAuthServiceService_IRefreshAccessToken;
}

interface IAuthServiceService_IRegisterByPassword extends grpc.MethodDefinition<auth_auth_pb.RegisterByPasswordRequest, auth_auth_pb.AuthResponse> {
    path: "/auth.AuthService/RegisterByPassword";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.RegisterByPasswordRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.RegisterByPasswordRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.AuthResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.AuthResponse>;
}
interface IAuthServiceService_ILoginByPassword extends grpc.MethodDefinition<auth_auth_pb.LoginByPasswordRequest, auth_auth_pb.AuthResponse> {
    path: "/auth.AuthService/LoginByPassword";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.LoginByPasswordRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.LoginByPasswordRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.AuthResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.AuthResponse>;
}
interface IAuthServiceService_IRefreshAccessToken extends grpc.MethodDefinition<auth_auth_pb.RefreshAccessTokenRequest, auth_auth_pb.AuthResponse> {
    path: "/auth.AuthService/RefreshAccessToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.RefreshAccessTokenRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.RefreshAccessTokenRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.AuthResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.AuthResponse>;
}

export const AuthServiceService: IAuthServiceService;

export interface IAuthServiceServer extends grpc.UntypedServiceImplementation {
    registerByPassword: grpc.handleUnaryCall<auth_auth_pb.RegisterByPasswordRequest, auth_auth_pb.AuthResponse>;
    loginByPassword: grpc.handleUnaryCall<auth_auth_pb.LoginByPasswordRequest, auth_auth_pb.AuthResponse>;
    refreshAccessToken: grpc.handleUnaryCall<auth_auth_pb.RefreshAccessTokenRequest, auth_auth_pb.AuthResponse>;
}

export interface IAuthServiceClient {
    registerByPassword(request: auth_auth_pb.RegisterByPasswordRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    registerByPassword(request: auth_auth_pb.RegisterByPasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    registerByPassword(request: auth_auth_pb.RegisterByPasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    loginByPassword(request: auth_auth_pb.LoginByPasswordRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    loginByPassword(request: auth_auth_pb.LoginByPasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    loginByPassword(request: auth_auth_pb.LoginByPasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    refreshAccessToken(request: auth_auth_pb.RefreshAccessTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    refreshAccessToken(request: auth_auth_pb.RefreshAccessTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    refreshAccessToken(request: auth_auth_pb.RefreshAccessTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
}

export class AuthServiceClient extends grpc.Client implements IAuthServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public registerByPassword(request: auth_auth_pb.RegisterByPasswordRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    public registerByPassword(request: auth_auth_pb.RegisterByPasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    public registerByPassword(request: auth_auth_pb.RegisterByPasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    public loginByPassword(request: auth_auth_pb.LoginByPasswordRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    public loginByPassword(request: auth_auth_pb.LoginByPasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    public loginByPassword(request: auth_auth_pb.LoginByPasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    public refreshAccessToken(request: auth_auth_pb.RefreshAccessTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    public refreshAccessToken(request: auth_auth_pb.RefreshAccessTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
    public refreshAccessToken(request: auth_auth_pb.RefreshAccessTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthResponse) => void): grpc.ClientUnaryCall;
}
