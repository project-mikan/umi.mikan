// package: auth
// file: auth/auth_password.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as auth_auth_password_pb from "../auth/auth_password_pb";

interface IAuthServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    registerByPassword: IAuthServiceService_IRegisterByPassword;
    loginByPassword: IAuthServiceService_ILoginByPassword;
}

interface IAuthServiceService_IRegisterByPassword extends grpc.MethodDefinition<auth_auth_password_pb.RegisterByPasswordRequest, auth_auth_password_pb.RegisterResponse> {
    path: "/auth.AuthService/RegisterByPassword";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_password_pb.RegisterByPasswordRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_password_pb.RegisterByPasswordRequest>;
    responseSerialize: grpc.serialize<auth_auth_password_pb.RegisterResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_password_pb.RegisterResponse>;
}
interface IAuthServiceService_ILoginByPassword extends grpc.MethodDefinition<auth_auth_password_pb.LoginByPasswordRequest, auth_auth_password_pb.LoginResponse> {
    path: "/auth.AuthService/LoginByPassword";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_password_pb.LoginByPasswordRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_password_pb.LoginByPasswordRequest>;
    responseSerialize: grpc.serialize<auth_auth_password_pb.LoginResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_password_pb.LoginResponse>;
}

export const AuthServiceService: IAuthServiceService;

export interface IAuthServiceServer extends grpc.UntypedServiceImplementation {
    registerByPassword: grpc.handleUnaryCall<auth_auth_password_pb.RegisterByPasswordRequest, auth_auth_password_pb.RegisterResponse>;
    loginByPassword: grpc.handleUnaryCall<auth_auth_password_pb.LoginByPasswordRequest, auth_auth_password_pb.LoginResponse>;
}

export interface IAuthServiceClient {
    registerByPassword(request: auth_auth_password_pb.RegisterByPasswordRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.RegisterResponse) => void): grpc.ClientUnaryCall;
    registerByPassword(request: auth_auth_password_pb.RegisterByPasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.RegisterResponse) => void): grpc.ClientUnaryCall;
    registerByPassword(request: auth_auth_password_pb.RegisterByPasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.RegisterResponse) => void): grpc.ClientUnaryCall;
    loginByPassword(request: auth_auth_password_pb.LoginByPasswordRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.LoginResponse) => void): grpc.ClientUnaryCall;
    loginByPassword(request: auth_auth_password_pb.LoginByPasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.LoginResponse) => void): grpc.ClientUnaryCall;
    loginByPassword(request: auth_auth_password_pb.LoginByPasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.LoginResponse) => void): grpc.ClientUnaryCall;
}

export class AuthServiceClient extends grpc.Client implements IAuthServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public registerByPassword(request: auth_auth_password_pb.RegisterByPasswordRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.RegisterResponse) => void): grpc.ClientUnaryCall;
    public registerByPassword(request: auth_auth_password_pb.RegisterByPasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.RegisterResponse) => void): grpc.ClientUnaryCall;
    public registerByPassword(request: auth_auth_password_pb.RegisterByPasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.RegisterResponse) => void): grpc.ClientUnaryCall;
    public loginByPassword(request: auth_auth_password_pb.LoginByPasswordRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.LoginResponse) => void): grpc.ClientUnaryCall;
    public loginByPassword(request: auth_auth_password_pb.LoginByPasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.LoginResponse) => void): grpc.ClientUnaryCall;
    public loginByPassword(request: auth_auth_password_pb.LoginByPasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_password_pb.LoginResponse) => void): grpc.ClientUnaryCall;
}
