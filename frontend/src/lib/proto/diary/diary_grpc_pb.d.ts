// package: diary
// file: diary/diary.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as diary_diary_pb from "../diary/diary_pb";

interface IDiaryServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    createDiaryEntry: IDiaryServiceService_ICreateDiaryEntry;
    getDiaryEntry: IDiaryServiceService_IGetDiaryEntry;
    listDiaryEntries: IDiaryServiceService_IListDiaryEntries;
    searchDiaryEntries: IDiaryServiceService_ISearchDiaryEntries;
    updateDiaryEntry: IDiaryServiceService_IUpdateDiaryEntry;
    deleteDiaryEntry: IDiaryServiceService_IDeleteDiaryEntry;
}

interface IDiaryServiceService_ICreateDiaryEntry extends grpc.MethodDefinition<diary_diary_pb.CreateDiaryEntryRequest, diary_diary_pb.CreateDiaryEntryResponse> {
    path: "/diary.DiaryService/CreateDiaryEntry";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<diary_diary_pb.CreateDiaryEntryRequest>;
    requestDeserialize: grpc.deserialize<diary_diary_pb.CreateDiaryEntryRequest>;
    responseSerialize: grpc.serialize<diary_diary_pb.CreateDiaryEntryResponse>;
    responseDeserialize: grpc.deserialize<diary_diary_pb.CreateDiaryEntryResponse>;
}
interface IDiaryServiceService_IGetDiaryEntry extends grpc.MethodDefinition<diary_diary_pb.GetDiaryEntryRequest, diary_diary_pb.GetDiaryEntryResponse> {
    path: "/diary.DiaryService/GetDiaryEntry";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<diary_diary_pb.GetDiaryEntryRequest>;
    requestDeserialize: grpc.deserialize<diary_diary_pb.GetDiaryEntryRequest>;
    responseSerialize: grpc.serialize<diary_diary_pb.GetDiaryEntryResponse>;
    responseDeserialize: grpc.deserialize<diary_diary_pb.GetDiaryEntryResponse>;
}
interface IDiaryServiceService_IListDiaryEntries extends grpc.MethodDefinition<diary_diary_pb.ListDiaryEntriesRequest, diary_diary_pb.ListDiaryEntriesResponse> {
    path: "/diary.DiaryService/ListDiaryEntries";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<diary_diary_pb.ListDiaryEntriesRequest>;
    requestDeserialize: grpc.deserialize<diary_diary_pb.ListDiaryEntriesRequest>;
    responseSerialize: grpc.serialize<diary_diary_pb.ListDiaryEntriesResponse>;
    responseDeserialize: grpc.deserialize<diary_diary_pb.ListDiaryEntriesResponse>;
}
interface IDiaryServiceService_ISearchDiaryEntries extends grpc.MethodDefinition<diary_diary_pb.SearchDiaryEntriesRequest, diary_diary_pb.SearchDiaryEntriesResponse> {
    path: "/diary.DiaryService/SearchDiaryEntries";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<diary_diary_pb.SearchDiaryEntriesRequest>;
    requestDeserialize: grpc.deserialize<diary_diary_pb.SearchDiaryEntriesRequest>;
    responseSerialize: grpc.serialize<diary_diary_pb.SearchDiaryEntriesResponse>;
    responseDeserialize: grpc.deserialize<diary_diary_pb.SearchDiaryEntriesResponse>;
}
interface IDiaryServiceService_IUpdateDiaryEntry extends grpc.MethodDefinition<diary_diary_pb.UpdateDiaryEntryRequest, diary_diary_pb.UpdateDiaryEntryResponse> {
    path: "/diary.DiaryService/UpdateDiaryEntry";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<diary_diary_pb.UpdateDiaryEntryRequest>;
    requestDeserialize: grpc.deserialize<diary_diary_pb.UpdateDiaryEntryRequest>;
    responseSerialize: grpc.serialize<diary_diary_pb.UpdateDiaryEntryResponse>;
    responseDeserialize: grpc.deserialize<diary_diary_pb.UpdateDiaryEntryResponse>;
}
interface IDiaryServiceService_IDeleteDiaryEntry extends grpc.MethodDefinition<diary_diary_pb.DeleteDiaryEntryRequest, diary_diary_pb.DeleteDiaryEntryResponse> {
    path: "/diary.DiaryService/DeleteDiaryEntry";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<diary_diary_pb.DeleteDiaryEntryRequest>;
    requestDeserialize: grpc.deserialize<diary_diary_pb.DeleteDiaryEntryRequest>;
    responseSerialize: grpc.serialize<diary_diary_pb.DeleteDiaryEntryResponse>;
    responseDeserialize: grpc.deserialize<diary_diary_pb.DeleteDiaryEntryResponse>;
}

export const DiaryServiceService: IDiaryServiceService;

export interface IDiaryServiceServer extends grpc.UntypedServiceImplementation {
    createDiaryEntry: grpc.handleUnaryCall<diary_diary_pb.CreateDiaryEntryRequest, diary_diary_pb.CreateDiaryEntryResponse>;
    getDiaryEntry: grpc.handleUnaryCall<diary_diary_pb.GetDiaryEntryRequest, diary_diary_pb.GetDiaryEntryResponse>;
    listDiaryEntries: grpc.handleUnaryCall<diary_diary_pb.ListDiaryEntriesRequest, diary_diary_pb.ListDiaryEntriesResponse>;
    searchDiaryEntries: grpc.handleUnaryCall<diary_diary_pb.SearchDiaryEntriesRequest, diary_diary_pb.SearchDiaryEntriesResponse>;
    updateDiaryEntry: grpc.handleUnaryCall<diary_diary_pb.UpdateDiaryEntryRequest, diary_diary_pb.UpdateDiaryEntryResponse>;
    deleteDiaryEntry: grpc.handleUnaryCall<diary_diary_pb.DeleteDiaryEntryRequest, diary_diary_pb.DeleteDiaryEntryResponse>;
}

export interface IDiaryServiceClient {
    createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    listDiaryEntries(request: diary_diary_pb.ListDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    listDiaryEntries(request: diary_diary_pb.ListDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    listDiaryEntries(request: diary_diary_pb.ListDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
}

export class DiaryServiceClient extends grpc.Client implements IDiaryServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public listDiaryEntries(request: diary_diary_pb.ListDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public listDiaryEntries(request: diary_diary_pb.ListDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public listDiaryEntries(request: diary_diary_pb.ListDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
}
