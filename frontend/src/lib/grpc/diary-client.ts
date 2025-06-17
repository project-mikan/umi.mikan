import { create } from "@bufbuild/protobuf";
import { diaryClient } from "./client.js";
import {
	CreateDiaryEntryRequestSchema,
	type CreateDiaryEntryResponse,
	DeleteDiaryEntryRequestSchema,
	type DeleteDiaryEntryResponse,
	type DiaryEntry,
	GetDiaryEntriesByMonthRequestSchema,
	type GetDiaryEntriesByMonthResponse,
	GetDiaryEntriesRequestSchema,
	type GetDiaryEntriesResponse,
	GetDiaryEntryRequestSchema,
	type GetDiaryEntryResponse,
	SearchDiaryEntriesRequestSchema,
	type SearchDiaryEntriesResponse,
	UpdateDiaryEntryRequestSchema,
	type UpdateDiaryEntryResponse,
	type YM,
	type YMD,
	YMDSchema,
	YMSchema,
} from "./diary/diary_pb.js";

export interface CreateDiaryEntryParams {
	content: string;
	date: YMD;
}

export interface UpdateDiaryEntryParams {
	id: string;
	title: string;
	content: string;
	date: YMD;
}

export interface SearchDiaryEntriesParams {
	userID: string;
	keyword: string;
}

export class DiaryClient {
	async createDiaryEntry(
		params: CreateDiaryEntryParams,
	): Promise<CreateDiaryEntryResponse> {
		const request = create(CreateDiaryEntryRequestSchema, {
			content: params.content,
			date: params.date,
		});

		const response = await diaryClient.createDiaryEntry(request);
		return response;
	}

	async getDiaryEntry(date: YMD): Promise<GetDiaryEntryResponse> {
		const request = create(GetDiaryEntryRequestSchema, {
			date,
		});

		const response = await diaryClient.getDiaryEntry(request);
		return response;
	}

	async getDiaryEntries(dates: YMD[]): Promise<GetDiaryEntriesResponse> {
		const request = create(GetDiaryEntriesRequestSchema, {
			dates,
		});

		const response = await diaryClient.getDiaryEntries(request);
		return response;
	}

	async getDiaryEntriesByMonth(
		month: YM,
	): Promise<GetDiaryEntriesByMonthResponse> {
		const request = create(GetDiaryEntriesByMonthRequestSchema, {
			month,
		});

		const response = await diaryClient.getDiaryEntriesByMonth(request);
		return response;
	}

	async updateDiaryEntry(
		params: UpdateDiaryEntryParams,
	): Promise<UpdateDiaryEntryResponse> {
		const request = create(UpdateDiaryEntryRequestSchema, {
			id: params.id,
			title: params.title,
			content: params.content,
			date: params.date,
		});

		const response = await diaryClient.updateDiaryEntry(request);
		return response;
	}

	async deleteDiaryEntry(id: string): Promise<DeleteDiaryEntryResponse> {
		const request = create(DeleteDiaryEntryRequestSchema, {
			id,
		});

		const response = await diaryClient.deleteDiaryEntry(request);
		return response;
	}

	async searchDiaryEntries(
		params: SearchDiaryEntriesParams,
	): Promise<SearchDiaryEntriesResponse> {
		const request = create(SearchDiaryEntriesRequestSchema, {
			userID: params.userID,
			keyword: params.keyword,
		});

		const response = await diaryClient.searchDiaryEntries(request);
		return response;
	}
}

export const diaryService = new DiaryClient();

export function createYMD(year: number, month: number, day: number): YMD {
	return create(YMDSchema, { year, month, day });
}

export function createYM(year: number, month: number): YM {
	return create(YMSchema, { year, month });
}
