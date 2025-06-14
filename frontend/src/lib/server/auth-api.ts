import { create } from '@bufbuild/protobuf';
import { createClient } from '@connectrpc/connect';
import { createGrpcWebTransport } from '@connectrpc/connect-web';
import { 
  AuthService, 
  LoginByPasswordRequestSchema, 
  RegisterByPasswordRequestSchema,
  type AuthResponse
} from '$lib/grpc/auth/auth_pb.js';

const transport = createGrpcWebTransport({
  baseUrl: 'http://backend:8080'
});

const authClient = createClient(AuthService, transport);

export interface LoginByPasswordParams {
  email: string;
  password: string;
}

export interface RegisterByPasswordParams {
  email: string;
  password: string;
  name: string;
}

export async function loginByPassword(params: LoginByPasswordParams): Promise<AuthResponse> {
  const request = create(LoginByPasswordRequestSchema, {
    email: params.email,
    password: params.password
  });

  const response = await authClient.loginByPassword(request);
  return response;
}

export async function registerByPassword(params: RegisterByPasswordParams): Promise<AuthResponse> {
  const request = create(RegisterByPasswordRequestSchema, {
    email: params.email,
    password: params.password,
    name: params.name
  });

  const response = await authClient.registerByPassword(request);
  return response;
}