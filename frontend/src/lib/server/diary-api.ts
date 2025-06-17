import { create } from '@bufbuild/protobuf';
import { createClient } from '@connectrpc/connect';
import { createGrpcWebTransport } from '@connectrpc/connect-web';
import { 
  DiaryService,
  CreateDiaryEntryRequestSchema,
  GetDiaryEntryRequestSchema,
  GetDiaryEntriesByMonthRequestSchema,
  UpdateDiaryEntryRequestSchema,
  DeleteDiaryEntryRequestSchema,
  YMDSchema,
  YMSchema,
  type CreateDiaryEntryResponse,
  type GetDiaryEntryResponse,
  type GetDiaryEntriesByMonthResponse,
  type UpdateDiaryEntryResponse,
  type DeleteDiaryEntryResponse,
  type YMD,
  type YM
} from '$lib/grpc/diary/diary_pb.js';

function createAuthenticatedTransport(accessToken: string) {
  return createGrpcWebTransport({
    baseUrl: 'http://backend:8080',
    interceptors: [
      (next) => (req) => {
        req.header.set('authorization', `Bearer ${accessToken}`);
        return next(req);
      }
    ]
  });
}

export interface CreateDiaryEntryParams {
  content: string;
  date: YMD;
  accessToken: string;
}

export interface GetDiaryEntryParams {
  date: YMD;
  accessToken: string;
}

export interface GetDiaryEntriesByMonthParams {
  month: YM;
  accessToken: string;
}

export interface UpdateDiaryEntryParams {
  id: string;
  title: string;
  content: string;
  date: YMD;
  accessToken: string;
}

export interface DeleteDiaryEntryParams {
  id: string;
  accessToken: string;
}

export async function createDiaryEntry(params: CreateDiaryEntryParams): Promise<CreateDiaryEntryResponse> {
  const transport = createAuthenticatedTransport(params.accessToken);
  const client = createClient(DiaryService, transport);
  
  const request = create(CreateDiaryEntryRequestSchema, {
    content: params.content,
    date: params.date
  });

  return await client.createDiaryEntry(request);
}

export async function getDiaryEntry(params: GetDiaryEntryParams): Promise<GetDiaryEntryResponse> {
  const transport = createAuthenticatedTransport(params.accessToken);
  const client = createClient(DiaryService, transport);
  
  const request = create(GetDiaryEntryRequestSchema, {
    date: params.date
  });

  return await client.getDiaryEntry(request);
}

export async function getDiaryEntriesByMonth(params: GetDiaryEntriesByMonthParams): Promise<GetDiaryEntriesByMonthResponse> {
  const transport = createAuthenticatedTransport(params.accessToken);
  const client = createClient(DiaryService, transport);
  
  const request = create(GetDiaryEntriesByMonthRequestSchema, {
    month: params.month
  });

  return await client.getDiaryEntriesByMonth(request);
}

export async function updateDiaryEntry(params: UpdateDiaryEntryParams): Promise<UpdateDiaryEntryResponse> {
  const transport = createAuthenticatedTransport(params.accessToken);
  const client = createClient(DiaryService, transport);
  
  const request = create(UpdateDiaryEntryRequestSchema, {
    id: params.id,
    title: params.title,
    content: params.content,
    date: params.date
  });

  return await client.updateDiaryEntry(request);
}

export async function deleteDiaryEntry(params: DeleteDiaryEntryParams): Promise<DeleteDiaryEntryResponse> {
  const transport = createAuthenticatedTransport(params.accessToken);
  const client = createClient(DiaryService, transport);
  
  const request = create(DeleteDiaryEntryRequestSchema, {
    id: params.id
  });

  return await client.deleteDiaryEntry(request);
}

export function createYMD(year: number, month: number, day: number): YMD {
  return create(YMDSchema, { year, month, day });
}

export function createYM(year: number, month: number): YM {
  return create(YMSchema, { year, month });
}