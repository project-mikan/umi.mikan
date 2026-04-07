import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import {
  CreateDiaryEntryRequestSchema,
  type CreateDiaryEntryResponse,
  DeleteDiaryEntryRequestSchema,
  type DeleteDiaryEntryResponse,
  DiaryService,
  GenerateDailySummaryRequestSchema,
  type GenerateDailySummaryResponse,
  GenerateMonthlySummaryRequestSchema,
  type GenerateMonthlySummaryResponse,
  GetDailySummaryRequestSchema,
  type GetDailySummaryResponse,
  GetDiaryEntriesByMonthRequestSchema,
  type GetDiaryEntriesByMonthResponse,
  GetDiaryEntryRequestSchema,
  type GetDiaryEntryResponse,
  GetLatestTrendRequestSchema,
  type GetLatestTrendResponse,
  GetMonthlySummaryRequestSchema,
  type GetMonthlySummaryResponse,
  SearchDiaryEntriesRequestSchema,
  type SearchDiaryEntriesResponse,
  SearchDiaryEntriesSemanticRequestSchema,
  type SearchDiaryEntriesSemanticResponse,
  TriggerLatestTrendRequestSchema,
  type TriggerLatestTrendResponse,
  UpdateDiaryEntryRequestSchema,
  type UpdateDiaryEntryResponse,
  type YM,
  type YMD,
  YMDSchema,
  YMSchema,
  TriggerDiaryHighlightRequestSchema,
  type TriggerDiaryHighlightResponse,
  GetDiaryHighlightRequestSchema,
  type GetDiaryHighlightResponse,
  RegenerateAllEmbeddingsRequestSchema,
  type RegenerateAllEmbeddingsResponse,
  GetDiaryEmbeddingStatusRequestSchema,
  type GetDiaryEmbeddingStatusResponse,
} from "$lib/grpc/diary/diary_pb";

// モジュールレベルで Transport と Client を共有する（リクエストごとの生成コストを排除）
// 認証ヘッダは各 RPC 呼び出し時に CallOptions として個別に付与する
const transport = createGrpcTransport({
  baseUrl: "http://backend:8080",
});

const diaryClient = createClient(DiaryService, transport);

function authHeader(accessToken: string) {
  return { headers: { authorization: `Bearer ${accessToken}` } };
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

export interface SearchDiaryEntriesParams {
  keyword: string;
  accessToken: string;
}

export interface GenerateMonthlySummaryParams {
  month: YM;
  accessToken: string;
}

export interface GetMonthlySummaryParams {
  month: YM;
  accessToken: string;
}

export async function createDiaryEntry(
  params: CreateDiaryEntryParams,
): Promise<CreateDiaryEntryResponse> {
  const request = create(CreateDiaryEntryRequestSchema, {
    content: params.content,
    date: params.date,
  });

  return await diaryClient.createDiaryEntry(
    request,
    authHeader(params.accessToken),
  );
}

export async function getDiaryEntry(
  params: GetDiaryEntryParams,
): Promise<GetDiaryEntryResponse> {
  const request = create(GetDiaryEntryRequestSchema, {
    date: params.date,
  });

  return await diaryClient.getDiaryEntry(
    request,
    authHeader(params.accessToken),
  );
}

export async function getDiaryEntriesByMonth(
  params: GetDiaryEntriesByMonthParams,
): Promise<GetDiaryEntriesByMonthResponse> {
  const request = create(GetDiaryEntriesByMonthRequestSchema, {
    month: params.month,
  });

  return await diaryClient.getDiaryEntriesByMonth(
    request,
    authHeader(params.accessToken),
  );
}

export async function updateDiaryEntry(
  params: UpdateDiaryEntryParams,
): Promise<UpdateDiaryEntryResponse> {
  const request = create(UpdateDiaryEntryRequestSchema, {
    id: params.id,
    title: params.title,
    content: params.content,
    date: params.date,
  });

  return await diaryClient.updateDiaryEntry(
    request,
    authHeader(params.accessToken),
  );
}

export async function deleteDiaryEntry(
  params: DeleteDiaryEntryParams,
): Promise<DeleteDiaryEntryResponse> {
  const request = create(DeleteDiaryEntryRequestSchema, {
    id: params.id,
  });

  return await diaryClient.deleteDiaryEntry(
    request,
    authHeader(params.accessToken),
  );
}

export function createYMD(year: number, month: number, day: number): YMD {
  return create(YMDSchema, { year, month, day });
}

export function createYM(year: number, month: number): YM {
  return create(YMSchema, { year, month });
}

export async function searchDiaryEntries(
  params: SearchDiaryEntriesParams,
): Promise<SearchDiaryEntriesResponse> {
  const request = create(SearchDiaryEntriesRequestSchema, {
    keyword: params.keyword,
  });

  return await diaryClient.searchDiaryEntries(
    request,
    authHeader(params.accessToken),
  );
}

export async function generateMonthlySummary(
  params: GenerateMonthlySummaryParams,
): Promise<GenerateMonthlySummaryResponse> {
  const request = create(GenerateMonthlySummaryRequestSchema, {
    month: params.month,
  });

  return await diaryClient.generateMonthlySummary(
    request,
    authHeader(params.accessToken),
  );
}

export async function getMonthlySummary(
  params: GetMonthlySummaryParams,
): Promise<GetMonthlySummaryResponse> {
  const request = create(GetMonthlySummaryRequestSchema, {
    month: params.month,
  });

  return await diaryClient.getMonthlySummary(
    request,
    authHeader(params.accessToken),
  );
}

export interface GenerateDailySummaryParams {
  diaryId: string;
  accessToken: string;
}

export interface GetDailySummaryParams {
  date: YMD;
  accessToken: string;
}

export async function generateDailySummary(
  params: GenerateDailySummaryParams,
): Promise<GenerateDailySummaryResponse> {
  const request = create(GenerateDailySummaryRequestSchema, {
    diaryId: params.diaryId,
  });

  return await diaryClient.generateDailySummary(
    request,
    authHeader(params.accessToken),
  );
}

export async function getDailySummary(
  params: GetDailySummaryParams,
): Promise<GetDailySummaryResponse> {
  const request = create(GetDailySummaryRequestSchema, {
    date: params.date,
  });

  return await diaryClient.getDailySummary(
    request,
    authHeader(params.accessToken),
  );
}

export interface GetLatestTrendParams {
  accessToken: string;
}

export interface TriggerLatestTrendParams {
  accessToken: string;
}

export async function getLatestTrend(
  params: GetLatestTrendParams,
): Promise<GetLatestTrendResponse> {
  const request = create(GetLatestTrendRequestSchema, {});

  return await diaryClient.getLatestTrend(
    request,
    authHeader(params.accessToken),
  );
}

export async function triggerLatestTrend(
  params: TriggerLatestTrendParams,
): Promise<TriggerLatestTrendResponse> {
  const request = create(TriggerLatestTrendRequestSchema, {});

  return await diaryClient.triggerLatestTrend(
    request,
    authHeader(params.accessToken),
  );
}

export interface TriggerDiaryHighlightParams {
  diaryId: string;
  accessToken: string;
}

export interface GetDiaryHighlightParams {
  diaryId: string;
  accessToken: string;
}

export async function triggerDiaryHighlight(
  params: TriggerDiaryHighlightParams,
): Promise<TriggerDiaryHighlightResponse> {
  const request = create(TriggerDiaryHighlightRequestSchema, {
    diaryId: params.diaryId,
  });

  return await diaryClient.triggerDiaryHighlight(
    request,
    authHeader(params.accessToken),
  );
}

export async function getDiaryHighlight(
  params: GetDiaryHighlightParams,
): Promise<GetDiaryHighlightResponse> {
  const request = create(GetDiaryHighlightRequestSchema, {
    diaryId: params.diaryId,
  });

  return await diaryClient.getDiaryHighlight(
    request,
    authHeader(params.accessToken),
  );
}

export interface SearchDiaryEntriesSemanticParams {
  query: string;
  limit?: number;
  accessToken: string;
}

export async function searchDiaryEntriesSemantic(
  params: SearchDiaryEntriesSemanticParams,
): Promise<SearchDiaryEntriesSemanticResponse> {
  const request = create(SearchDiaryEntriesSemanticRequestSchema, {
    query: params.query,
    limit: params.limit ?? 10,
  });

  return await diaryClient.searchDiaryEntriesSemantic(
    request,
    authHeader(params.accessToken),
  );
}

export interface RegenerateAllEmbeddingsParams {
  accessToken: string;
}

export async function regenerateAllEmbeddings(
  params: RegenerateAllEmbeddingsParams,
): Promise<RegenerateAllEmbeddingsResponse> {
  const request = create(RegenerateAllEmbeddingsRequestSchema, {});
  return await diaryClient.regenerateAllEmbeddings(
    request,
    authHeader(params.accessToken),
  );
}

export interface GetDiaryEmbeddingStatusParams {
  diaryId: string;
  accessToken: string;
}

export async function getDiaryEmbeddingStatus(
  params: GetDiaryEmbeddingStatusParams,
): Promise<GetDiaryEmbeddingStatusResponse> {
  const request = create(GetDiaryEmbeddingStatusRequestSchema, {
    diaryId: params.diaryId,
  });
  return await diaryClient.getDiaryEmbeddingStatus(
    request,
    authHeader(params.accessToken),
  );
}
