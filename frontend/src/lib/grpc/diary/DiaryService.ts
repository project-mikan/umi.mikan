// Original file: proto/diary/diary.proto

import type * as grpc from '@grpc/grpc-js'
import type { MethodDefinition } from '@grpc/proto-loader'
import type { CreateDiaryEntryRequest as _diary_CreateDiaryEntryRequest, CreateDiaryEntryRequest__Output as _diary_CreateDiaryEntryRequest__Output } from '../diary/CreateDiaryEntryRequest';
import type { CreateDiaryEntryResponse as _diary_CreateDiaryEntryResponse, CreateDiaryEntryResponse__Output as _diary_CreateDiaryEntryResponse__Output } from '../diary/CreateDiaryEntryResponse';
import type { DeleteDiaryEntryRequest as _diary_DeleteDiaryEntryRequest, DeleteDiaryEntryRequest__Output as _diary_DeleteDiaryEntryRequest__Output } from '../diary/DeleteDiaryEntryRequest';
import type { DeleteDiaryEntryResponse as _diary_DeleteDiaryEntryResponse, DeleteDiaryEntryResponse__Output as _diary_DeleteDiaryEntryResponse__Output } from '../diary/DeleteDiaryEntryResponse';
import type { GetDiaryEntriesByMonthRequest as _diary_GetDiaryEntriesByMonthRequest, GetDiaryEntriesByMonthRequest__Output as _diary_GetDiaryEntriesByMonthRequest__Output } from '../diary/GetDiaryEntriesByMonthRequest';
import type { GetDiaryEntriesByMonthResponse as _diary_GetDiaryEntriesByMonthResponse, GetDiaryEntriesByMonthResponse__Output as _diary_GetDiaryEntriesByMonthResponse__Output } from '../diary/GetDiaryEntriesByMonthResponse';
import type { GetDiaryEntriesRequest as _diary_GetDiaryEntriesRequest, GetDiaryEntriesRequest__Output as _diary_GetDiaryEntriesRequest__Output } from '../diary/GetDiaryEntriesRequest';
import type { GetDiaryEntriesResponse as _diary_GetDiaryEntriesResponse, GetDiaryEntriesResponse__Output as _diary_GetDiaryEntriesResponse__Output } from '../diary/GetDiaryEntriesResponse';
import type { GetDiaryEntryRequest as _diary_GetDiaryEntryRequest, GetDiaryEntryRequest__Output as _diary_GetDiaryEntryRequest__Output } from '../diary/GetDiaryEntryRequest';
import type { GetDiaryEntryResponse as _diary_GetDiaryEntryResponse, GetDiaryEntryResponse__Output as _diary_GetDiaryEntryResponse__Output } from '../diary/GetDiaryEntryResponse';
import type { SearchDiaryEntriesRequest as _diary_SearchDiaryEntriesRequest, SearchDiaryEntriesRequest__Output as _diary_SearchDiaryEntriesRequest__Output } from '../diary/SearchDiaryEntriesRequest';
import type { SearchDiaryEntriesResponse as _diary_SearchDiaryEntriesResponse, SearchDiaryEntriesResponse__Output as _diary_SearchDiaryEntriesResponse__Output } from '../diary/SearchDiaryEntriesResponse';
import type { UpdateDiaryEntryRequest as _diary_UpdateDiaryEntryRequest, UpdateDiaryEntryRequest__Output as _diary_UpdateDiaryEntryRequest__Output } from '../diary/UpdateDiaryEntryRequest';
import type { UpdateDiaryEntryResponse as _diary_UpdateDiaryEntryResponse, UpdateDiaryEntryResponse__Output as _diary_UpdateDiaryEntryResponse__Output } from '../diary/UpdateDiaryEntryResponse';

export interface DiaryServiceClient extends grpc.Client {
  CreateDiaryEntry(argument: _diary_CreateDiaryEntryRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_CreateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  CreateDiaryEntry(argument: _diary_CreateDiaryEntryRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_CreateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  CreateDiaryEntry(argument: _diary_CreateDiaryEntryRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_CreateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  CreateDiaryEntry(argument: _diary_CreateDiaryEntryRequest, callback: grpc.requestCallback<_diary_CreateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  createDiaryEntry(argument: _diary_CreateDiaryEntryRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_CreateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  createDiaryEntry(argument: _diary_CreateDiaryEntryRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_CreateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  createDiaryEntry(argument: _diary_CreateDiaryEntryRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_CreateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  createDiaryEntry(argument: _diary_CreateDiaryEntryRequest, callback: grpc.requestCallback<_diary_CreateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  
  DeleteDiaryEntry(argument: _diary_DeleteDiaryEntryRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_DeleteDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  DeleteDiaryEntry(argument: _diary_DeleteDiaryEntryRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_DeleteDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  DeleteDiaryEntry(argument: _diary_DeleteDiaryEntryRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_DeleteDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  DeleteDiaryEntry(argument: _diary_DeleteDiaryEntryRequest, callback: grpc.requestCallback<_diary_DeleteDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  deleteDiaryEntry(argument: _diary_DeleteDiaryEntryRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_DeleteDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  deleteDiaryEntry(argument: _diary_DeleteDiaryEntryRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_DeleteDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  deleteDiaryEntry(argument: _diary_DeleteDiaryEntryRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_DeleteDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  deleteDiaryEntry(argument: _diary_DeleteDiaryEntryRequest, callback: grpc.requestCallback<_diary_DeleteDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  
  GetDiaryEntries(argument: _diary_GetDiaryEntriesRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  GetDiaryEntries(argument: _diary_GetDiaryEntriesRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_GetDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  GetDiaryEntries(argument: _diary_GetDiaryEntriesRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  GetDiaryEntries(argument: _diary_GetDiaryEntriesRequest, callback: grpc.requestCallback<_diary_GetDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntries(argument: _diary_GetDiaryEntriesRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntries(argument: _diary_GetDiaryEntriesRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_GetDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntries(argument: _diary_GetDiaryEntriesRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntries(argument: _diary_GetDiaryEntriesRequest, callback: grpc.requestCallback<_diary_GetDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  
  GetDiaryEntriesByMonth(argument: _diary_GetDiaryEntriesByMonthRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntriesByMonthResponse__Output>): grpc.ClientUnaryCall;
  GetDiaryEntriesByMonth(argument: _diary_GetDiaryEntriesByMonthRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_GetDiaryEntriesByMonthResponse__Output>): grpc.ClientUnaryCall;
  GetDiaryEntriesByMonth(argument: _diary_GetDiaryEntriesByMonthRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntriesByMonthResponse__Output>): grpc.ClientUnaryCall;
  GetDiaryEntriesByMonth(argument: _diary_GetDiaryEntriesByMonthRequest, callback: grpc.requestCallback<_diary_GetDiaryEntriesByMonthResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntriesByMonth(argument: _diary_GetDiaryEntriesByMonthRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntriesByMonthResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntriesByMonth(argument: _diary_GetDiaryEntriesByMonthRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_GetDiaryEntriesByMonthResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntriesByMonth(argument: _diary_GetDiaryEntriesByMonthRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntriesByMonthResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntriesByMonth(argument: _diary_GetDiaryEntriesByMonthRequest, callback: grpc.requestCallback<_diary_GetDiaryEntriesByMonthResponse__Output>): grpc.ClientUnaryCall;
  
  GetDiaryEntry(argument: _diary_GetDiaryEntryRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  GetDiaryEntry(argument: _diary_GetDiaryEntryRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_GetDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  GetDiaryEntry(argument: _diary_GetDiaryEntryRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  GetDiaryEntry(argument: _diary_GetDiaryEntryRequest, callback: grpc.requestCallback<_diary_GetDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntry(argument: _diary_GetDiaryEntryRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntry(argument: _diary_GetDiaryEntryRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_GetDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntry(argument: _diary_GetDiaryEntryRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_GetDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  getDiaryEntry(argument: _diary_GetDiaryEntryRequest, callback: grpc.requestCallback<_diary_GetDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  
  SearchDiaryEntries(argument: _diary_SearchDiaryEntriesRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_SearchDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  SearchDiaryEntries(argument: _diary_SearchDiaryEntriesRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_SearchDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  SearchDiaryEntries(argument: _diary_SearchDiaryEntriesRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_SearchDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  SearchDiaryEntries(argument: _diary_SearchDiaryEntriesRequest, callback: grpc.requestCallback<_diary_SearchDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  searchDiaryEntries(argument: _diary_SearchDiaryEntriesRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_SearchDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  searchDiaryEntries(argument: _diary_SearchDiaryEntriesRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_SearchDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  searchDiaryEntries(argument: _diary_SearchDiaryEntriesRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_SearchDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  searchDiaryEntries(argument: _diary_SearchDiaryEntriesRequest, callback: grpc.requestCallback<_diary_SearchDiaryEntriesResponse__Output>): grpc.ClientUnaryCall;
  
  UpdateDiaryEntry(argument: _diary_UpdateDiaryEntryRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_UpdateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  UpdateDiaryEntry(argument: _diary_UpdateDiaryEntryRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_UpdateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  UpdateDiaryEntry(argument: _diary_UpdateDiaryEntryRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_UpdateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  UpdateDiaryEntry(argument: _diary_UpdateDiaryEntryRequest, callback: grpc.requestCallback<_diary_UpdateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  updateDiaryEntry(argument: _diary_UpdateDiaryEntryRequest, metadata: grpc.Metadata, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_UpdateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  updateDiaryEntry(argument: _diary_UpdateDiaryEntryRequest, metadata: grpc.Metadata, callback: grpc.requestCallback<_diary_UpdateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  updateDiaryEntry(argument: _diary_UpdateDiaryEntryRequest, options: grpc.CallOptions, callback: grpc.requestCallback<_diary_UpdateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  updateDiaryEntry(argument: _diary_UpdateDiaryEntryRequest, callback: grpc.requestCallback<_diary_UpdateDiaryEntryResponse__Output>): grpc.ClientUnaryCall;
  
}

export interface DiaryServiceHandlers extends grpc.UntypedServiceImplementation {
  CreateDiaryEntry: grpc.handleUnaryCall<_diary_CreateDiaryEntryRequest__Output, _diary_CreateDiaryEntryResponse>;
  
  DeleteDiaryEntry: grpc.handleUnaryCall<_diary_DeleteDiaryEntryRequest__Output, _diary_DeleteDiaryEntryResponse>;
  
  GetDiaryEntries: grpc.handleUnaryCall<_diary_GetDiaryEntriesRequest__Output, _diary_GetDiaryEntriesResponse>;
  
  GetDiaryEntriesByMonth: grpc.handleUnaryCall<_diary_GetDiaryEntriesByMonthRequest__Output, _diary_GetDiaryEntriesByMonthResponse>;
  
  GetDiaryEntry: grpc.handleUnaryCall<_diary_GetDiaryEntryRequest__Output, _diary_GetDiaryEntryResponse>;
  
  SearchDiaryEntries: grpc.handleUnaryCall<_diary_SearchDiaryEntriesRequest__Output, _diary_SearchDiaryEntriesResponse>;
  
  UpdateDiaryEntry: grpc.handleUnaryCall<_diary_UpdateDiaryEntryRequest__Output, _diary_UpdateDiaryEntryResponse>;
  
}

export interface DiaryServiceDefinition extends grpc.ServiceDefinition {
  CreateDiaryEntry: MethodDefinition<_diary_CreateDiaryEntryRequest, _diary_CreateDiaryEntryResponse, _diary_CreateDiaryEntryRequest__Output, _diary_CreateDiaryEntryResponse__Output>
  DeleteDiaryEntry: MethodDefinition<_diary_DeleteDiaryEntryRequest, _diary_DeleteDiaryEntryResponse, _diary_DeleteDiaryEntryRequest__Output, _diary_DeleteDiaryEntryResponse__Output>
  GetDiaryEntries: MethodDefinition<_diary_GetDiaryEntriesRequest, _diary_GetDiaryEntriesResponse, _diary_GetDiaryEntriesRequest__Output, _diary_GetDiaryEntriesResponse__Output>
  GetDiaryEntriesByMonth: MethodDefinition<_diary_GetDiaryEntriesByMonthRequest, _diary_GetDiaryEntriesByMonthResponse, _diary_GetDiaryEntriesByMonthRequest__Output, _diary_GetDiaryEntriesByMonthResponse__Output>
  GetDiaryEntry: MethodDefinition<_diary_GetDiaryEntryRequest, _diary_GetDiaryEntryResponse, _diary_GetDiaryEntryRequest__Output, _diary_GetDiaryEntryResponse__Output>
  SearchDiaryEntries: MethodDefinition<_diary_SearchDiaryEntriesRequest, _diary_SearchDiaryEntriesResponse, _diary_SearchDiaryEntriesRequest__Output, _diary_SearchDiaryEntriesResponse__Output>
  UpdateDiaryEntry: MethodDefinition<_diary_UpdateDiaryEntryRequest, _diary_UpdateDiaryEntryResponse, _diary_UpdateDiaryEntryRequest__Output, _diary_UpdateDiaryEntryResponse__Output>
}
