// package: diary
// file: diary/diary.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "grpc";
import * as diary_diary_pb from "../diary/diary_pb";

interface IDiaryServiceService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    createDiaryEntry: IDiaryServiceService_ICreateDiaryEntry;
    updateDiaryEntry: IDiaryServiceService_IUpdateDiaryEntry;
    deleteDiaryEntry: IDiaryServiceService_IDeleteDiaryEntry;
    getDiaryEntry: IDiaryServiceService_IGetDiaryEntry;
    getDiaryEntries: IDiaryServiceService_IGetDiaryEntries;
    getDiaryEntriesByMonth: IDiaryServiceService_IGetDiaryEntriesByMonth;
    searchDiaryEntries: IDiaryServiceService_ISearchDiaryEntries;
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
interface IDiaryServiceService_IGetDiaryEntry extends grpc.MethodDefinition<diary_diary_pb.GetDiaryEntryRequest, diary_diary_pb.GetDiaryEntryResponse> {
    path: "/diary.DiaryService/GetDiaryEntry";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<diary_diary_pb.GetDiaryEntryRequest>;
    requestDeserialize: grpc.deserialize<diary_diary_pb.GetDiaryEntryRequest>;
    responseSerialize: grpc.serialize<diary_diary_pb.GetDiaryEntryResponse>;
    responseDeserialize: grpc.deserialize<diary_diary_pb.GetDiaryEntryResponse>;
}
interface IDiaryServiceService_IGetDiaryEntries extends grpc.MethodDefinition<diary_diary_pb.GetDiaryEntriesRequest, diary_diary_pb.GetDiaryEntriesResponse> {
    path: "/diary.DiaryService/GetDiaryEntries";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<diary_diary_pb.GetDiaryEntriesRequest>;
    requestDeserialize: grpc.deserialize<diary_diary_pb.GetDiaryEntriesRequest>;
    responseSerialize: grpc.serialize<diary_diary_pb.GetDiaryEntriesResponse>;
    responseDeserialize: grpc.deserialize<diary_diary_pb.GetDiaryEntriesResponse>;
}
interface IDiaryServiceService_IGetDiaryEntriesByMonth extends grpc.MethodDefinition<diary_diary_pb.GetDiaryEntriesByMonthRequest, diary_diary_pb.GetDiaryEntriesByMonthResponse> {
    path: "/diary.DiaryService/GetDiaryEntriesByMonth";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<diary_diary_pb.GetDiaryEntriesByMonthRequest>;
    requestDeserialize: grpc.deserialize<diary_diary_pb.GetDiaryEntriesByMonthRequest>;
    responseSerialize: grpc.serialize<diary_diary_pb.GetDiaryEntriesByMonthResponse>;
    responseDeserialize: grpc.deserialize<diary_diary_pb.GetDiaryEntriesByMonthResponse>;
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

export const DiaryServiceService: IDiaryServiceService;

export interface IDiaryServiceServer {
    createDiaryEntry: grpc.handleUnaryCall<diary_diary_pb.CreateDiaryEntryRequest, diary_diary_pb.CreateDiaryEntryResponse>;
    updateDiaryEntry: grpc.handleUnaryCall<diary_diary_pb.UpdateDiaryEntryRequest, diary_diary_pb.UpdateDiaryEntryResponse>;
    deleteDiaryEntry: grpc.handleUnaryCall<diary_diary_pb.DeleteDiaryEntryRequest, diary_diary_pb.DeleteDiaryEntryResponse>;
    getDiaryEntry: grpc.handleUnaryCall<diary_diary_pb.GetDiaryEntryRequest, diary_diary_pb.GetDiaryEntryResponse>;
    getDiaryEntries: grpc.handleUnaryCall<diary_diary_pb.GetDiaryEntriesRequest, diary_diary_pb.GetDiaryEntriesResponse>;
    getDiaryEntriesByMonth: grpc.handleUnaryCall<diary_diary_pb.GetDiaryEntriesByMonthRequest, diary_diary_pb.GetDiaryEntriesByMonthResponse>;
    searchDiaryEntries: grpc.handleUnaryCall<diary_diary_pb.SearchDiaryEntriesRequest, diary_diary_pb.SearchDiaryEntriesResponse>;
}

export interface IDiaryServiceClient {
    createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntries(request: diary_diary_pb.GetDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntries(request: diary_diary_pb.GetDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntries(request: diary_diary_pb.GetDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntriesByMonth(request: diary_diary_pb.GetDiaryEntriesByMonthRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesByMonthResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntriesByMonth(request: diary_diary_pb.GetDiaryEntriesByMonthRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesByMonthResponse) => void): grpc.ClientUnaryCall;
    getDiaryEntriesByMonth(request: diary_diary_pb.GetDiaryEntriesByMonthRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesByMonthResponse) => void): grpc.ClientUnaryCall;
    searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
}

export class DiaryServiceClient extends grpc.Client implements IDiaryServiceClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
    public createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public createDiaryEntry(request: diary_diary_pb.CreateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.CreateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public updateDiaryEntry(request: diary_diary_pb.UpdateDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.UpdateDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public deleteDiaryEntry(request: diary_diary_pb.DeleteDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.DeleteDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntry(request: diary_diary_pb.GetDiaryEntryRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntryResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntries(request: diary_diary_pb.GetDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntries(request: diary_diary_pb.GetDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntries(request: diary_diary_pb.GetDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntriesByMonth(request: diary_diary_pb.GetDiaryEntriesByMonthRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesByMonthResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntriesByMonth(request: diary_diary_pb.GetDiaryEntriesByMonthRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesByMonthResponse) => void): grpc.ClientUnaryCall;
    public getDiaryEntriesByMonth(request: diary_diary_pb.GetDiaryEntriesByMonthRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.GetDiaryEntriesByMonthResponse) => void): grpc.ClientUnaryCall;
    public searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
    public searchDiaryEntries(request: diary_diary_pb.SearchDiaryEntriesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: diary_diary_pb.SearchDiaryEntriesResponse) => void): grpc.ClientUnaryCall;
}
