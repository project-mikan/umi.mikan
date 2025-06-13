// package: auth
// file: auth/auth.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class RefreshAccessTokenRequest extends jspb.Message { 
    getRefreshToken(): string;
    setRefreshToken(value: string): RefreshAccessTokenRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RefreshAccessTokenRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RefreshAccessTokenRequest): RefreshAccessTokenRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RefreshAccessTokenRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RefreshAccessTokenRequest;
    static deserializeBinaryFromReader(message: RefreshAccessTokenRequest, reader: jspb.BinaryReader): RefreshAccessTokenRequest;
}

export namespace RefreshAccessTokenRequest {
    export type AsObject = {
        refreshToken: string,
    }
}

export class RegisterByPasswordRequest extends jspb.Message { 
    getEmail(): string;
    setEmail(value: string): RegisterByPasswordRequest;
    getPassword(): string;
    setPassword(value: string): RegisterByPasswordRequest;
    getName(): string;
    setName(value: string): RegisterByPasswordRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RegisterByPasswordRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RegisterByPasswordRequest): RegisterByPasswordRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RegisterByPasswordRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RegisterByPasswordRequest;
    static deserializeBinaryFromReader(message: RegisterByPasswordRequest, reader: jspb.BinaryReader): RegisterByPasswordRequest;
}

export namespace RegisterByPasswordRequest {
    export type AsObject = {
        email: string,
        password: string,
        name: string,
    }
}

export class LoginByPasswordRequest extends jspb.Message { 
    getEmail(): string;
    setEmail(value: string): LoginByPasswordRequest;
    getPassword(): string;
    setPassword(value: string): LoginByPasswordRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): LoginByPasswordRequest.AsObject;
    static toObject(includeInstance: boolean, msg: LoginByPasswordRequest): LoginByPasswordRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: LoginByPasswordRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): LoginByPasswordRequest;
    static deserializeBinaryFromReader(message: LoginByPasswordRequest, reader: jspb.BinaryReader): LoginByPasswordRequest;
}

export namespace LoginByPasswordRequest {
    export type AsObject = {
        email: string,
        password: string,
    }
}

export class AuthResponse extends jspb.Message { 
    getAccessToken(): string;
    setAccessToken(value: string): AuthResponse;
    getTokenType(): string;
    setTokenType(value: string): AuthResponse;
    getExpiresIn(): number;
    setExpiresIn(value: number): AuthResponse;
    getRefreshToken(): string;
    setRefreshToken(value: string): AuthResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AuthResponse.AsObject;
    static toObject(includeInstance: boolean, msg: AuthResponse): AuthResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AuthResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AuthResponse;
    static deserializeBinaryFromReader(message: AuthResponse, reader: jspb.BinaryReader): AuthResponse;
}

export namespace AuthResponse {
    export type AsObject = {
        accessToken: string,
        tokenType: string,
        expiresIn: number,
        refreshToken: string,
    }
}
