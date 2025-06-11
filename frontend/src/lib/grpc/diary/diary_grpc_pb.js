// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var diary_diary_pb = require('../diary/diary_pb.js');

function serialize_diary_CreateDiaryEntryRequest(arg) {
  if (!(arg instanceof diary_diary_pb.CreateDiaryEntryRequest)) {
    throw new Error('Expected argument of type diary.CreateDiaryEntryRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_CreateDiaryEntryRequest(buffer_arg) {
  return diary_diary_pb.CreateDiaryEntryRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_CreateDiaryEntryResponse(arg) {
  if (!(arg instanceof diary_diary_pb.CreateDiaryEntryResponse)) {
    throw new Error('Expected argument of type diary.CreateDiaryEntryResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_CreateDiaryEntryResponse(buffer_arg) {
  return diary_diary_pb.CreateDiaryEntryResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_DeleteDiaryEntryRequest(arg) {
  if (!(arg instanceof diary_diary_pb.DeleteDiaryEntryRequest)) {
    throw new Error('Expected argument of type diary.DeleteDiaryEntryRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_DeleteDiaryEntryRequest(buffer_arg) {
  return diary_diary_pb.DeleteDiaryEntryRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_DeleteDiaryEntryResponse(arg) {
  if (!(arg instanceof diary_diary_pb.DeleteDiaryEntryResponse)) {
    throw new Error('Expected argument of type diary.DeleteDiaryEntryResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_DeleteDiaryEntryResponse(buffer_arg) {
  return diary_diary_pb.DeleteDiaryEntryResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_GetDiaryEntriesByMonthRequest(arg) {
  if (!(arg instanceof diary_diary_pb.GetDiaryEntriesByMonthRequest)) {
    throw new Error('Expected argument of type diary.GetDiaryEntriesByMonthRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_GetDiaryEntriesByMonthRequest(buffer_arg) {
  return diary_diary_pb.GetDiaryEntriesByMonthRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_GetDiaryEntriesByMonthResponse(arg) {
  if (!(arg instanceof diary_diary_pb.GetDiaryEntriesByMonthResponse)) {
    throw new Error('Expected argument of type diary.GetDiaryEntriesByMonthResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_GetDiaryEntriesByMonthResponse(buffer_arg) {
  return diary_diary_pb.GetDiaryEntriesByMonthResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_GetDiaryEntriesRequest(arg) {
  if (!(arg instanceof diary_diary_pb.GetDiaryEntriesRequest)) {
    throw new Error('Expected argument of type diary.GetDiaryEntriesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_GetDiaryEntriesRequest(buffer_arg) {
  return diary_diary_pb.GetDiaryEntriesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_GetDiaryEntriesResponse(arg) {
  if (!(arg instanceof diary_diary_pb.GetDiaryEntriesResponse)) {
    throw new Error('Expected argument of type diary.GetDiaryEntriesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_GetDiaryEntriesResponse(buffer_arg) {
  return diary_diary_pb.GetDiaryEntriesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_GetDiaryEntryRequest(arg) {
  if (!(arg instanceof diary_diary_pb.GetDiaryEntryRequest)) {
    throw new Error('Expected argument of type diary.GetDiaryEntryRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_GetDiaryEntryRequest(buffer_arg) {
  return diary_diary_pb.GetDiaryEntryRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_GetDiaryEntryResponse(arg) {
  if (!(arg instanceof diary_diary_pb.GetDiaryEntryResponse)) {
    throw new Error('Expected argument of type diary.GetDiaryEntryResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_GetDiaryEntryResponse(buffer_arg) {
  return diary_diary_pb.GetDiaryEntryResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_SearchDiaryEntriesRequest(arg) {
  if (!(arg instanceof diary_diary_pb.SearchDiaryEntriesRequest)) {
    throw new Error('Expected argument of type diary.SearchDiaryEntriesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_SearchDiaryEntriesRequest(buffer_arg) {
  return diary_diary_pb.SearchDiaryEntriesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_SearchDiaryEntriesResponse(arg) {
  if (!(arg instanceof diary_diary_pb.SearchDiaryEntriesResponse)) {
    throw new Error('Expected argument of type diary.SearchDiaryEntriesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_SearchDiaryEntriesResponse(buffer_arg) {
  return diary_diary_pb.SearchDiaryEntriesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_UpdateDiaryEntryRequest(arg) {
  if (!(arg instanceof diary_diary_pb.UpdateDiaryEntryRequest)) {
    throw new Error('Expected argument of type diary.UpdateDiaryEntryRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_UpdateDiaryEntryRequest(buffer_arg) {
  return diary_diary_pb.UpdateDiaryEntryRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_diary_UpdateDiaryEntryResponse(arg) {
  if (!(arg instanceof diary_diary_pb.UpdateDiaryEntryResponse)) {
    throw new Error('Expected argument of type diary.UpdateDiaryEntryResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_diary_UpdateDiaryEntryResponse(buffer_arg) {
  return diary_diary_pb.UpdateDiaryEntryResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var DiaryServiceService = exports.DiaryServiceService = {
  // 作成
createDiaryEntry: {
    path: '/diary.DiaryService/CreateDiaryEntry',
    requestStream: false,
    responseStream: false,
    requestType: diary_diary_pb.CreateDiaryEntryRequest,
    responseType: diary_diary_pb.CreateDiaryEntryResponse,
    requestSerialize: serialize_diary_CreateDiaryEntryRequest,
    requestDeserialize: deserialize_diary_CreateDiaryEntryRequest,
    responseSerialize: serialize_diary_CreateDiaryEntryResponse,
    responseDeserialize: deserialize_diary_CreateDiaryEntryResponse,
  },
  // 更新
updateDiaryEntry: {
    path: '/diary.DiaryService/UpdateDiaryEntry',
    requestStream: false,
    responseStream: false,
    requestType: diary_diary_pb.UpdateDiaryEntryRequest,
    responseType: diary_diary_pb.UpdateDiaryEntryResponse,
    requestSerialize: serialize_diary_UpdateDiaryEntryRequest,
    requestDeserialize: deserialize_diary_UpdateDiaryEntryRequest,
    responseSerialize: serialize_diary_UpdateDiaryEntryResponse,
    responseDeserialize: deserialize_diary_UpdateDiaryEntryResponse,
  },
  // 削除
deleteDiaryEntry: {
    path: '/diary.DiaryService/DeleteDiaryEntry',
    requestStream: false,
    responseStream: false,
    requestType: diary_diary_pb.DeleteDiaryEntryRequest,
    responseType: diary_diary_pb.DeleteDiaryEntryResponse,
    requestSerialize: serialize_diary_DeleteDiaryEntryRequest,
    requestDeserialize: deserialize_diary_DeleteDiaryEntryRequest,
    responseSerialize: serialize_diary_DeleteDiaryEntryResponse,
    responseDeserialize: deserialize_diary_DeleteDiaryEntryResponse,
  },
  // 日付指定で単体取得
getDiaryEntry: {
    path: '/diary.DiaryService/GetDiaryEntry',
    requestStream: false,
    responseStream: false,
    requestType: diary_diary_pb.GetDiaryEntryRequest,
    responseType: diary_diary_pb.GetDiaryEntryResponse,
    requestSerialize: serialize_diary_GetDiaryEntryRequest,
    requestDeserialize: deserialize_diary_GetDiaryEntryRequest,
    responseSerialize: serialize_diary_GetDiaryEntryResponse,
    responseDeserialize: deserialize_diary_GetDiaryEntryResponse,
  },
  // 日付指定で複数取得(ホームでの表示などで直近3日とかほしいケースや過去数年分ほしいケースに対応)
getDiaryEntries: {
    path: '/diary.DiaryService/GetDiaryEntries',
    requestStream: false,
    responseStream: false,
    requestType: diary_diary_pb.GetDiaryEntriesRequest,
    responseType: diary_diary_pb.GetDiaryEntriesResponse,
    requestSerialize: serialize_diary_GetDiaryEntriesRequest,
    requestDeserialize: deserialize_diary_GetDiaryEntriesRequest,
    responseSerialize: serialize_diary_GetDiaryEntriesResponse,
    responseDeserialize: deserialize_diary_GetDiaryEntriesResponse,
  },
  // 月ごとに取得
getDiaryEntriesByMonth: {
    path: '/diary.DiaryService/GetDiaryEntriesByMonth',
    requestStream: false,
    responseStream: false,
    requestType: diary_diary_pb.GetDiaryEntriesByMonthRequest,
    responseType: diary_diary_pb.GetDiaryEntriesByMonthResponse,
    requestSerialize: serialize_diary_GetDiaryEntriesByMonthRequest,
    requestDeserialize: deserialize_diary_GetDiaryEntriesByMonthRequest,
    responseSerialize: serialize_diary_GetDiaryEntriesByMonthResponse,
    responseDeserialize: deserialize_diary_GetDiaryEntriesByMonthResponse,
  },
  // 検索
searchDiaryEntries: {
    path: '/diary.DiaryService/SearchDiaryEntries',
    requestStream: false,
    responseStream: false,
    requestType: diary_diary_pb.SearchDiaryEntriesRequest,
    responseType: diary_diary_pb.SearchDiaryEntriesResponse,
    requestSerialize: serialize_diary_SearchDiaryEntriesRequest,
    requestDeserialize: deserialize_diary_SearchDiaryEntriesRequest,
    responseSerialize: serialize_diary_SearchDiaryEntriesResponse,
    responseDeserialize: deserialize_diary_SearchDiaryEntriesResponse,
  },
};

exports.DiaryServiceClient = grpc.makeGenericClientConstructor(DiaryServiceService, 'DiaryService');
