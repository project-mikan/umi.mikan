// package: auth
// file: auth/auth_password.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class RegisterByPasswordRequest extends jspb.Message { 
    getEmail(): string;
    setEmail(value: string): RegisterByPasswordRequest;
    getPassword(): string;
    setPassword(value: string): RegisterByPasswordRequest;

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

export class RegisterResponse extends jspb.Message { 
    getJwt(): string;
    setJwt(value: string): RegisterResponse;
    getMessage(): string;
    setMessage(value: string): RegisterResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RegisterResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RegisterResponse): RegisterResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RegisterResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RegisterResponse;
    static deserializeBinaryFromReader(message: RegisterResponse, reader: jspb.BinaryReader): RegisterResponse;
}

export namespace RegisterResponse {
    export type AsObject = {
        jwt: string,
        message: string,
    }
}

export class LoginResponse extends jspb.Message { 
    getJwt(): string;
    setJwt(value: string): LoginResponse;
    getMessage(): string;
    setMessage(value: string): LoginResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): LoginResponse.AsObject;
    static toObject(includeInstance: boolean, msg: LoginResponse): LoginResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: LoginResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): LoginResponse;
    static deserializeBinaryFromReader(message: LoginResponse, reader: jspb.BinaryReader): LoginResponse;
}

export namespace LoginResponse {
    export type AsObject = {
        jwt: string,
        message: string,
    }
}
