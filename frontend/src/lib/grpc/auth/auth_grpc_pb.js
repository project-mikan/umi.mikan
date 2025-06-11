// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var auth_auth_pb = require('../auth/auth_pb.js');

function serialize_auth_AuthResponse(arg) {
  if (!(arg instanceof auth_auth_pb.AuthResponse)) {
    throw new Error('Expected argument of type auth.AuthResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_AuthResponse(buffer_arg) {
  return auth_auth_pb.AuthResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_LoginByPasswordRequest(arg) {
  if (!(arg instanceof auth_auth_pb.LoginByPasswordRequest)) {
    throw new Error('Expected argument of type auth.LoginByPasswordRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_LoginByPasswordRequest(buffer_arg) {
  return auth_auth_pb.LoginByPasswordRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RefreshAccessTokenRequest(arg) {
  if (!(arg instanceof auth_auth_pb.RefreshAccessTokenRequest)) {
    throw new Error('Expected argument of type auth.RefreshAccessTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RefreshAccessTokenRequest(buffer_arg) {
  return auth_auth_pb.RefreshAccessTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RegisterByPasswordRequest(arg) {
  if (!(arg instanceof auth_auth_pb.RegisterByPasswordRequest)) {
    throw new Error('Expected argument of type auth.RegisterByPasswordRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RegisterByPasswordRequest(buffer_arg) {
  return auth_auth_pb.RegisterByPasswordRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


var AuthServiceService = exports.AuthServiceService = {
  // 新規登録
registerByPassword: {
    path: '/auth.AuthService/RegisterByPassword',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.RegisterByPasswordRequest,
    responseType: auth_auth_pb.AuthResponse,
    requestSerialize: serialize_auth_RegisterByPasswordRequest,
    requestDeserialize: deserialize_auth_RegisterByPasswordRequest,
    responseSerialize: serialize_auth_AuthResponse,
    responseDeserialize: deserialize_auth_AuthResponse,
  },
  // ログイン
loginByPassword: {
    path: '/auth.AuthService/LoginByPassword',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.LoginByPasswordRequest,
    responseType: auth_auth_pb.AuthResponse,
    requestSerialize: serialize_auth_LoginByPasswordRequest,
    requestDeserialize: deserialize_auth_LoginByPasswordRequest,
    responseSerialize: serialize_auth_AuthResponse,
    responseDeserialize: deserialize_auth_AuthResponse,
  },
  // AccessTokenの更新
refreshAccessToken: {
    path: '/auth.AuthService/RefreshAccessToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.RefreshAccessTokenRequest,
    responseType: auth_auth_pb.AuthResponse,
    requestSerialize: serialize_auth_RefreshAccessTokenRequest,
    requestDeserialize: deserialize_auth_RefreshAccessTokenRequest,
    responseSerialize: serialize_auth_AuthResponse,
    responseDeserialize: deserialize_auth_AuthResponse,
  },
};

exports.AuthServiceClient = grpc.makeGenericClientConstructor(AuthServiceService, 'AuthService');
