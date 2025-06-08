// package: diary
// file: diary/diary.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class Date extends jspb.Message { 
    getYear(): number;
    setYear(value: number): Date;
    getMonth(): number;
    setMonth(value: number): Date;
    getDay(): number;
    setDay(value: number): Date;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Date.AsObject;
    static toObject(includeInstance: boolean, msg: Date): Date.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Date, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Date;
    static deserializeBinaryFromReader(message: Date, reader: jspb.BinaryReader): Date;
}

export namespace Date {
    export type AsObject = {
        year: number,
        month: number,
        day: number,
    }
}

export class DiaryEntry extends jspb.Message { 
    getId(): string;
    setId(value: string): DiaryEntry;
    getContent(): string;
    setContent(value: string): DiaryEntry;

    hasDate(): boolean;
    clearDate(): void;
    getDate(): Date | undefined;
    setDate(value?: Date): DiaryEntry;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DiaryEntry.AsObject;
    static toObject(includeInstance: boolean, msg: DiaryEntry): DiaryEntry.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DiaryEntry, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DiaryEntry;
    static deserializeBinaryFromReader(message: DiaryEntry, reader: jspb.BinaryReader): DiaryEntry;
}

export namespace DiaryEntry {
    export type AsObject = {
        id: string,
        content: string,
        date?: Date.AsObject,
    }
}

export class CreateDiaryEntryRequest extends jspb.Message { 
    getTitle(): string;
    setTitle(value: string): CreateDiaryEntryRequest;
    getContent(): string;
    setContent(value: string): CreateDiaryEntryRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateDiaryEntryRequest.AsObject;
    static toObject(includeInstance: boolean, msg: CreateDiaryEntryRequest): CreateDiaryEntryRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateDiaryEntryRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateDiaryEntryRequest;
    static deserializeBinaryFromReader(message: CreateDiaryEntryRequest, reader: jspb.BinaryReader): CreateDiaryEntryRequest;
}

export namespace CreateDiaryEntryRequest {
    export type AsObject = {
        title: string,
        content: string,
    }
}

export class CreateDiaryEntryResponse extends jspb.Message { 

    hasEntry(): boolean;
    clearEntry(): void;
    getEntry(): DiaryEntry | undefined;
    setEntry(value?: DiaryEntry): CreateDiaryEntryResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CreateDiaryEntryResponse.AsObject;
    static toObject(includeInstance: boolean, msg: CreateDiaryEntryResponse): CreateDiaryEntryResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CreateDiaryEntryResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CreateDiaryEntryResponse;
    static deserializeBinaryFromReader(message: CreateDiaryEntryResponse, reader: jspb.BinaryReader): CreateDiaryEntryResponse;
}

export namespace CreateDiaryEntryResponse {
    export type AsObject = {
        entry?: DiaryEntry.AsObject,
    }
}

export class GetDiaryEntryRequest extends jspb.Message { 
    getId(): string;
    setId(value: string): GetDiaryEntryRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetDiaryEntryRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetDiaryEntryRequest): GetDiaryEntryRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetDiaryEntryRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetDiaryEntryRequest;
    static deserializeBinaryFromReader(message: GetDiaryEntryRequest, reader: jspb.BinaryReader): GetDiaryEntryRequest;
}

export namespace GetDiaryEntryRequest {
    export type AsObject = {
        id: string,
    }
}

export class SearchDiaryEntriesRequest extends jspb.Message { 
    getUserid(): string;
    setUserid(value: string): SearchDiaryEntriesRequest;
    getKeyword(): string;
    setKeyword(value: string): SearchDiaryEntriesRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SearchDiaryEntriesRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SearchDiaryEntriesRequest): SearchDiaryEntriesRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SearchDiaryEntriesRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SearchDiaryEntriesRequest;
    static deserializeBinaryFromReader(message: SearchDiaryEntriesRequest, reader: jspb.BinaryReader): SearchDiaryEntriesRequest;
}

export namespace SearchDiaryEntriesRequest {
    export type AsObject = {
        userid: string,
        keyword: string,
    }
}

export class SearchDiaryEntriesResponse extends jspb.Message { 
    clearEntriesList(): void;
    getEntriesList(): Array<DiaryEntry>;
    setEntriesList(value: Array<DiaryEntry>): SearchDiaryEntriesResponse;
    addEntries(value?: DiaryEntry, index?: number): DiaryEntry;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SearchDiaryEntriesResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SearchDiaryEntriesResponse): SearchDiaryEntriesResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SearchDiaryEntriesResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SearchDiaryEntriesResponse;
    static deserializeBinaryFromReader(message: SearchDiaryEntriesResponse, reader: jspb.BinaryReader): SearchDiaryEntriesResponse;
}

export namespace SearchDiaryEntriesResponse {
    export type AsObject = {
        entriesList: Array<DiaryEntry.AsObject>,
    }
}

export class GetDiaryEntryResponse extends jspb.Message { 

    hasEntry(): boolean;
    clearEntry(): void;
    getEntry(): DiaryEntry | undefined;
    setEntry(value?: DiaryEntry): GetDiaryEntryResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetDiaryEntryResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetDiaryEntryResponse): GetDiaryEntryResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetDiaryEntryResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetDiaryEntryResponse;
    static deserializeBinaryFromReader(message: GetDiaryEntryResponse, reader: jspb.BinaryReader): GetDiaryEntryResponse;
}

export namespace GetDiaryEntryResponse {
    export type AsObject = {
        entry?: DiaryEntry.AsObject,
    }
}

export class ListDiaryEntriesRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListDiaryEntriesRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListDiaryEntriesRequest): ListDiaryEntriesRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListDiaryEntriesRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListDiaryEntriesRequest;
    static deserializeBinaryFromReader(message: ListDiaryEntriesRequest, reader: jspb.BinaryReader): ListDiaryEntriesRequest;
}

export namespace ListDiaryEntriesRequest {
    export type AsObject = {
    }
}

export class ListDiaryEntriesResponse extends jspb.Message { 
    clearEntriesList(): void;
    getEntriesList(): Array<DiaryEntry>;
    setEntriesList(value: Array<DiaryEntry>): ListDiaryEntriesResponse;
    addEntries(value?: DiaryEntry, index?: number): DiaryEntry;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListDiaryEntriesResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListDiaryEntriesResponse): ListDiaryEntriesResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListDiaryEntriesResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListDiaryEntriesResponse;
    static deserializeBinaryFromReader(message: ListDiaryEntriesResponse, reader: jspb.BinaryReader): ListDiaryEntriesResponse;
}

export namespace ListDiaryEntriesResponse {
    export type AsObject = {
        entriesList: Array<DiaryEntry.AsObject>,
    }
}

export class UpdateDiaryEntryRequest extends jspb.Message { 
    getId(): string;
    setId(value: string): UpdateDiaryEntryRequest;
    getTitle(): string;
    setTitle(value: string): UpdateDiaryEntryRequest;
    getContent(): string;
    setContent(value: string): UpdateDiaryEntryRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdateDiaryEntryRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UpdateDiaryEntryRequest): UpdateDiaryEntryRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdateDiaryEntryRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdateDiaryEntryRequest;
    static deserializeBinaryFromReader(message: UpdateDiaryEntryRequest, reader: jspb.BinaryReader): UpdateDiaryEntryRequest;
}

export namespace UpdateDiaryEntryRequest {
    export type AsObject = {
        id: string,
        title: string,
        content: string,
    }
}

export class UpdateDiaryEntryResponse extends jspb.Message { 

    hasEntry(): boolean;
    clearEntry(): void;
    getEntry(): DiaryEntry | undefined;
    setEntry(value?: DiaryEntry): UpdateDiaryEntryResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdateDiaryEntryResponse.AsObject;
    static toObject(includeInstance: boolean, msg: UpdateDiaryEntryResponse): UpdateDiaryEntryResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdateDiaryEntryResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdateDiaryEntryResponse;
    static deserializeBinaryFromReader(message: UpdateDiaryEntryResponse, reader: jspb.BinaryReader): UpdateDiaryEntryResponse;
}

export namespace UpdateDiaryEntryResponse {
    export type AsObject = {
        entry?: DiaryEntry.AsObject,
    }
}

export class DeleteDiaryEntryRequest extends jspb.Message { 
    getId(): string;
    setId(value: string): DeleteDiaryEntryRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteDiaryEntryRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteDiaryEntryRequest): DeleteDiaryEntryRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteDiaryEntryRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteDiaryEntryRequest;
    static deserializeBinaryFromReader(message: DeleteDiaryEntryRequest, reader: jspb.BinaryReader): DeleteDiaryEntryRequest;
}

export namespace DeleteDiaryEntryRequest {
    export type AsObject = {
        id: string,
    }
}

export class DeleteDiaryEntryResponse extends jspb.Message { 
    getSuccess(): boolean;
    setSuccess(value: boolean): DeleteDiaryEntryResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteDiaryEntryResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteDiaryEntryResponse): DeleteDiaryEntryResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteDiaryEntryResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteDiaryEntryResponse;
    static deserializeBinaryFromReader(message: DeleteDiaryEntryResponse, reader: jspb.BinaryReader): DeleteDiaryEntryResponse;
}

export namespace DeleteDiaryEntryResponse {
    export type AsObject = {
        success: boolean,
    }
}
