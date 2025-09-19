import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import {
	CreateDiaryEntryRequestSchema,
	type CreateDiaryEntryResponse,
	DeleteDiaryEntryRequestSchema,
	type DeleteDiaryEntryResponse,
	DiaryService,
	GenerateMonthlySummaryRequestSchema,
	type GenerateMonthlySummaryResponse,
	GetDiaryEntriesByMonthRequestSchema,
	type GetDiaryEntriesByMonthResponse,
	GetDiaryEntryRequestSchema,
	type GetDiaryEntryResponse,
	GetMonthlySummaryRequestSchema,
	type GetMonthlySummaryResponse,
	SearchDiaryEntriesRequestSchema,
	type SearchDiaryEntriesResponse,
	UpdateDiaryEntryRequestSchema,
	type UpdateDiaryEntryResponse,
	type YM,
	type YMD,
	YMDSchema,
	YMSchema,
} from "$lib/grpc/diary/diary_pb";

function createAuthenticatedTransport(accessToken: string) {
	return createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${accessToken}`);
				return next(req);
			},
		],
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
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(DiaryService, transport);

	const request = create(CreateDiaryEntryRequestSchema, {
		content: params.content,
		date: params.date,
	});

	return await client.createDiaryEntry(request);
}

export async function getDiaryEntry(
	params: GetDiaryEntryParams,
): Promise<GetDiaryEntryResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(DiaryService, transport);

	const request = create(GetDiaryEntryRequestSchema, {
		date: params.date,
	});

	return await client.getDiaryEntry(request);
}

export async function getDiaryEntriesByMonth(
	params: GetDiaryEntriesByMonthParams,
): Promise<GetDiaryEntriesByMonthResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(DiaryService, transport);

	const request = create(GetDiaryEntriesByMonthRequestSchema, {
		month: params.month,
	});

	return await client.getDiaryEntriesByMonth(request);
}

export async function updateDiaryEntry(
	params: UpdateDiaryEntryParams,
): Promise<UpdateDiaryEntryResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(DiaryService, transport);

	const request = create(UpdateDiaryEntryRequestSchema, {
		id: params.id,
		title: params.title,
		content: params.content,
		date: params.date,
	});

	return await client.updateDiaryEntry(request);
}

export async function deleteDiaryEntry(
	params: DeleteDiaryEntryParams,
): Promise<DeleteDiaryEntryResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(DiaryService, transport);

	const request = create(DeleteDiaryEntryRequestSchema, {
		id: params.id,
	});

	return await client.deleteDiaryEntry(request);
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
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(DiaryService, transport);

	const request = create(SearchDiaryEntriesRequestSchema, {
		keyword: params.keyword,
	});

	return await client.searchDiaryEntries(request);
}

export async function generateMonthlySummary(
	params: GenerateMonthlySummaryParams,
): Promise<GenerateMonthlySummaryResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(DiaryService, transport);

	const request = create(GenerateMonthlySummaryRequestSchema, {
		month: params.month,
	});

	return await client.generateMonthlySummary(request);
}

export async function getMonthlySummary(
	params: GetMonthlySummaryParams,
): Promise<GetMonthlySummaryResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(DiaryService, transport);

	const request = create(GetMonthlySummaryRequestSchema, {
		month: params.month,
	});

	return await client.getMonthlySummary(request);
}
