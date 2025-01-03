// package: diary
// file: hello.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as hello_pb from "./hello_pb";

interface IDiaryServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    createDiaryEntry: IDiaryServiceService_ICreateDiaryEntry;
    getDiaryEntry: IDiaryServiceService_IGetDiaryEntry;
    listDiaryEntries: IDiaryServiceService_IListDiaryEntries;
    searchDiaryEntries: IDiaryServiceService_ISearchDiaryEntries;
    updateDiaryEntry: IDiaryServiceService_IUpdateDiaryEntry;
    deleteDiaryEntry: IDiaryServiceService_IDeleteDiaryEntry;
}

interface IDiaryServiceService_ICreateDiaryEntry extends grpc.MethodDefinition<hello_pb.CreateDiaryEntryRequest, hello_pb.CreateDiaryEntryResponse> {
    path: "/diary.DiaryService/CreateDiaryEntry";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<hello_pb.CreateDiaryEntryRequest>;
    requestDeserialize: grpc.deserialize<hello_pb.CreateDiaryEntryRequest>;
    responseSerialize: grpc.serialize<hello_pb.CreateDiaryEntryResponse>;
    responseDeserialize: grpc.deserialize<hello_pb.CreateDiaryEntryResponse>;
}
interface IDiaryServiceService_IGetDiaryEntry extends grpc.MethodDefinition<hello_pb.GetDiaryEntryRequest, hello_pb.GetDiaryEntryResponse> {
    path: "/diary.DiaryService/GetDiaryEntry";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<hello_pb.GetDiaryEntryRequest>;
    requestDeserialize: grpc.deserialize<hello_pb.GetDiaryEntryRequest>;
    responseSerialize: grpc.serialize<hello_pb.GetDiaryEntryResponse>;
    responseDeserialize: grpc.deserialize<hello_pb.GetDiaryEntryResponse>;
}
interface IDiaryServiceService_IListDiaryEntries extends grpc.MethodDefinition<hello_pb.ListDiaryEntriesRequest, hello_pb.ListDiaryEntriesResponse> {
    path: "/diary.DiaryService/ListDiaryEntries";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<hello_pb.ListDiaryEntriesRequest>;
    requestDeserialize: grpc.deserialize<hello_pb.ListDiaryEntriesRequest>;
    responseSerialize: grpc.serialize<hello_pb.ListDiaryEntriesResponse>;
    responseDeserialize: grpc.deserialize<hello_pb.ListDiaryEntriesResponse>;
}
interface IDiaryServiceService_ISearchDiaryEntries extends grpc.MethodDefinition<hello_pb.SearchDiaryEntriesRequest, hello_pb.SearchDiaryEntriesResponse> {
    path: "/diary.DiaryService/SearchDiaryEntries";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<hello_pb.SearchDiaryEntriesRequest>;
    requestDeserialize: grpc.deserialize<hello_pb.SearchDiaryEntriesRequest>;
    responseSerialize: grpc.serialize<hello_pb.SearchDiaryEntriesResponse>;
    responseDeserialize: grpc.deserialize<hello_pb.SearchDiaryEntriesResponse>;
}
interface IDiaryServiceService_IUpdateDiaryEntry extends grpc.MethodDefinition<hello_pb.UpdateDiaryEntryRequest, hello_pb.UpdateDiaryEntryResponse> {
    path: "/diary.DiaryService/UpdateDiaryEntry";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<hello_pb.UpdateDiaryEntryRequest>;
    requestDeserialize: grpc.deserialize<hello_pb.UpdateDiaryEntryRequest>;
    responseSerialize: grpc.serialize<hello_pb.UpdateDiaryEntryResponse>;
    responseDeserialize: grpc.deserialize<hello_pb.UpdateDiaryEntryResponse>;
}
interface IDiaryServiceService_IDeleteDiaryEntry extends grpc.MethodDefinition<hello_pb.DeleteDiaryEntryRequest, hello_pb.DeleteDiaryEntryResponse> {
    path: "/diary.DiaryService/DeleteDiaryEntry";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<hello_pb.DeleteDiaryEntryRequest>;
    requestDeserialize: grpc.deserialize<hello_pb.DeleteDiaryEntryRequest>;
    responseSerialize: grpc.serialize<hello_pb.DeleteDiaryEntryResponse>;
    responseDeserialize: grpc.deserialize<hello_pb.DeleteDiaryEntryResponse>;
}

export const DiaryServiceService: IDiaryServiceService;

export interface IDiaryServiceServer extends grpc.UntypedServiceImplementation {
    createDiaryEntry: grpc.handleUnaryCall<hello_pb.CreateDiaryEntryRequest, hello_pb.CreateDiaryEntryResponse>;
    getDiaryEntry: grpc.handleUnaryCall<hello_pb.GetDiaryEntryRequest, hello_pb.GetDiaryEntryResponse>;
    listDiaryEntries: grpc.handleUnaryCall<hello_pb.ListDiaryEntriesRequest, hello_pb.ListDiaryEntriesResponse>;
    searchDiaryEntries: grpc.handleUnaryCall<hello_pb.SearchDiaryEntriesRequest, hello_pb.SearchDiaryEntriesResponse>;
    updateDiaryEntry: grpc.handleUnaryCall<hello_pb.UpdateDiaryEntryRequest, hello_pb.UpdateDiaryEntryResponse>;
    deleteDiaryEntry: grpc.handleUnaryCall<hello_pb.DeleteDiaryEntryRequest, hello_pb.DeleteDiaryEntryResponse>;
}

export interface IDiaryServiceClient {
    createDiaryEntry(request: hello_pb.CreateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    createDiaryEntry(request: hello_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    createDiaryEntry(request: hello_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntry(request: hello_pb.GetDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntry(request: hello_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntry(request: hello_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    listDiaryEntries(request: hello_pb.ListDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    listDiaryEntries(request: hello_pb.ListDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    listDiaryEntries(request: hello_pb.ListDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    searchDiaryEntries(request: hello_pb.SearchDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    searchDiaryEntries(request: hello_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    searchDiaryEntries(request: hello_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    updateDiaryEntry(request: hello_pb.UpdateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    updateDiaryEntry(request: hello_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    updateDiaryEntry(request: hello_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    deleteDiaryEntry(request: hello_pb.DeleteDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    deleteDiaryEntry(request: hello_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    deleteDiaryEntry(request: hello_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
}

export class DiaryServiceClient extends grpc.Client implements IDiaryServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public createDiaryEntry(request: hello_pb.CreateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public createDiaryEntry(request: hello_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public createDiaryEntry(request: hello_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntry(request: hello_pb.GetDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntry(request: hello_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntry(request: hello_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public listDiaryEntries(request: hello_pb.ListDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public listDiaryEntries(request: hello_pb.ListDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public listDiaryEntries(request: hello_pb.ListDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.ListDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public searchDiaryEntries(request: hello_pb.SearchDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public searchDiaryEntries(request: hello_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public searchDiaryEntries(request: hello_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public updateDiaryEntry(request: hello_pb.UpdateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public updateDiaryEntry(request: hello_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public updateDiaryEntry(request: hello_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public deleteDiaryEntry(request: hello_pb.DeleteDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: hello_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public deleteDiaryEntry(request: hello_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: hello_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public deleteDiaryEntry(request: hello_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: hello_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
}
