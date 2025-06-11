// package: diary
// file: diary/diary.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class YMD extends jspb.Message { 
    getYear(): number;
    setYear(value: number): YMD;
    getMonth(): number;
    setMonth(value: number): YMD;
    getDay(): number;
    setDay(value: number): YMD;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): YMD.AsObject;
    static toObject(includeInstance: boolean, msg: YMD): YMD.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: YMD, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): YMD;
    static deserializeBinaryFromReader(message: YMD, reader: jspb.BinaryReader): YMD;
}

export namespace YMD {
    export type AsObject = {
        year: number,
        month: number,
        day: number,
    }
}

export class YM extends jspb.Message { 
    getYear(): number;
    setYear(value: number): YM;
    getMonth(): number;
    setMonth(value: number): YM;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): YM.AsObject;
    static toObject(includeInstance: boolean, msg: YM): YM.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: YM, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): YM;
    static deserializeBinaryFromReader(message: YM, reader: jspb.BinaryReader): YM;
}

export namespace YM {
    export type AsObject = {
        year: number,
        month: number,
    }
}

export class DiaryEntry extends jspb.Message { 
    getId(): string;
    setId(value: string): DiaryEntry;

    hasDate(): boolean;
    clearDate(): void;
    getDate(): YMD | undefined;
    setDate(value?: YMD): DiaryEntry;
    getContent(): string;
    setContent(value: string): DiaryEntry;

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
        date?: YMD.AsObject,
        content: string,
    }
}

export class CreateDiaryEntryRequest extends jspb.Message { 
    getContent(): string;
    setContent(value: string): CreateDiaryEntryRequest;

    hasDate(): boolean;
    clearDate(): void;
    getDate(): YMD | undefined;
    setDate(value?: YMD): CreateDiaryEntryRequest;

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
        content: string,
        date?: YMD.AsObject,
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

    hasDate(): boolean;
    clearDate(): void;
    getDate(): YMD | undefined;
    setDate(value?: YMD): GetDiaryEntryRequest;

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
        date?: YMD.AsObject,
    }
}

export class GetDiaryEntriesRequest extends jspb.Message { 
    clearDatesList(): void;
    getDatesList(): Array<YMD>;
    setDatesList(value: Array<YMD>): GetDiaryEntriesRequest;
    addDates(value?: YMD, index?: number): YMD;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetDiaryEntriesRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetDiaryEntriesRequest): GetDiaryEntriesRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetDiaryEntriesRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetDiaryEntriesRequest;
    static deserializeBinaryFromReader(message: GetDiaryEntriesRequest, reader: jspb.BinaryReader): GetDiaryEntriesRequest;
}

export namespace GetDiaryEntriesRequest {
    export type AsObject = {
        datesList: Array<YMD.AsObject>,
    }
}

export class GetDiaryEntriesByMonthRequest extends jspb.Message { 

    hasMonth(): boolean;
    clearMonth(): void;
    getMonth(): YM | undefined;
    setMonth(value?: YM): GetDiaryEntriesByMonthRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetDiaryEntriesByMonthRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetDiaryEntriesByMonthRequest): GetDiaryEntriesByMonthRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetDiaryEntriesByMonthRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetDiaryEntriesByMonthRequest;
    static deserializeBinaryFromReader(message: GetDiaryEntriesByMonthRequest, reader: jspb.BinaryReader): GetDiaryEntriesByMonthRequest;
}

export namespace GetDiaryEntriesByMonthRequest {
    export type AsObject = {
        month?: YM.AsObject,
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
    getSearchedKeyword(): string;
    setSearchedKeyword(value: string): SearchDiaryEntriesResponse;
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
        searchedKeyword: string,
        entriesList: Array<DiaryEntry.AsObject>,
    }
}

export class GetDiaryEntriesResponse extends jspb.Message { 
    clearEntriesList(): void;
    getEntriesList(): Array<DiaryEntry>;
    setEntriesList(value: Array<DiaryEntry>): GetDiaryEntriesResponse;
    addEntries(value?: DiaryEntry, index?: number): DiaryEntry;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetDiaryEntriesResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetDiaryEntriesResponse): GetDiaryEntriesResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetDiaryEntriesResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetDiaryEntriesResponse;
    static deserializeBinaryFromReader(message: GetDiaryEntriesResponse, reader: jspb.BinaryReader): GetDiaryEntriesResponse;
}

export namespace GetDiaryEntriesResponse {
    export type AsObject = {
        entriesList: Array<DiaryEntry.AsObject>,
    }
}

export class GetDiaryEntriesByMonthResponse extends jspb.Message { 
    clearEntriesList(): void;
    getEntriesList(): Array<DiaryEntry>;
    setEntriesList(value: Array<DiaryEntry>): GetDiaryEntriesByMonthResponse;
    addEntries(value?: DiaryEntry, index?: number): DiaryEntry;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetDiaryEntriesByMonthResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetDiaryEntriesByMonthResponse): GetDiaryEntriesByMonthResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetDiaryEntriesByMonthResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetDiaryEntriesByMonthResponse;
    static deserializeBinaryFromReader(message: GetDiaryEntriesByMonthResponse, reader: jspb.BinaryReader): GetDiaryEntriesByMonthResponse;
}

export namespace GetDiaryEntriesByMonthResponse {
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

export class UpdateDiaryEntryRequest extends jspb.Message { 
    getId(): string;
    setId(value: string): UpdateDiaryEntryRequest;
    getTitle(): string;
    setTitle(value: string): UpdateDiaryEntryRequest;
    getContent(): string;
    setContent(value: string): UpdateDiaryEntryRequest;

    hasDate(): boolean;
    clearDate(): void;
    getDate(): YMD | undefined;
    setDate(value?: YMD): UpdateDiaryEntryRequest;

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
        date?: YMD.AsObject,
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
