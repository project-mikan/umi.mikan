// gRPC APIを使用したバックエンドとの通信（Diary Service）
import { diaryService } from './grpc-client';
import { create } from '@bufbuild/protobuf';
import { 
	YMDSchema,
	YMSchema,
	type YMD,
	type YM,
	type DiaryEntry
} from '../grpc/diary/diary_pb';

// Interface definitions
interface DateInput {
	year: number;
	month: number;
	day: number;
}

interface MonthInput {
	year: number;
	month: number;
}

interface CreateDiaryRequest {
	content: string;
	date: DateInput;
}

interface UpdateDiaryRequest {
	id: string;
	title: string;
	content: string;
	date: DateInput;
}

interface SearchRequest {
	userId: string;
	keyword: string;
}

interface DiaryEntryResult {
	id: string;
	date: DateInput;
	content: string;
}

// Helper function to create YMD message
function createYMD(date: DateInput): YMD {
	return create(YMDSchema, {
		year: date.year,
		month: date.month,
		day: date.day
	});
}

// Helper function to create YM message
function createYM(month: MonthInput): YM {
	return create(YMSchema, {
		year: month.year,
		month: month.month
	});
}

// Helper function to convert DiaryEntry to result
function convertDiaryEntry(entry: DiaryEntry): DiaryEntryResult {
	const date = entry.date;
	return {
		id: entry.id,
		date: {
			year: date?.year || 0,
			month: date?.month || 0,
			day: date?.day || 0
		},
		content: entry.content
	};
}

// API functions
export async function createDiaryEntry(request: CreateDiaryRequest): Promise<DiaryEntryResult> {
	try {
		const response = await diaryService.createDiaryEntry({
			content: request.content,
			date: createYMD(request.date)
		});

		const entry = response.entry;
		if (!entry) {
			throw new Error('No diary entry returned');
		}

		return convertDiaryEntry(entry);
	} catch (error) {
		// Handle specific gRPC errors
		if (error instanceof Error) {
			if (error.message.includes('409') || error.message.includes('ALREADY_EXISTS')) {
				throw new Error('Diary entry for this date already exists');
			} else if (error.message.includes('401') || error.message.includes('UNAUTHENTICATED')) {
				throw new Error('Authentication required');
			} else if (error.message.includes('503') || error.message.includes('UNAVAILABLE')) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to create diary entry: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function getDiaryEntry(date: DateInput): Promise<DiaryEntryResult | null> {
	try {
		const response = await diaryService.getDiaryEntry({
			date: createYMD(date)
		});

		const entry = response.entry;
		if (!entry) {
			return null;
		}

		return convertDiaryEntry(entry);
	} catch (error) {
		// Handle specific gRPC errors
		if (error instanceof Error) {
			if (error.message.includes('404') || error.message.includes('NOT_FOUND')) {
				return null;
			} else if (error.message.includes('401') || error.message.includes('UNAUTHENTICATED')) {
				throw new Error('Authentication required');
			} else if (error.message.includes('503') || error.message.includes('UNAVAILABLE')) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to get diary entry: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

// Additional diary service methods using simplified HTTP calls
async function callDiaryEndpoint(path: string, data: any): Promise<any> {
	const GRPC_SERVER_ADDRESS = process.env.GRPC_SERVER_ADDRESS || 'http://backend:8080';
	const response = await fetch(`${GRPC_SERVER_ADDRESS}${path}`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json',
			'Accept': 'application/json'
		},
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		throw new Error(`gRPC call failed: ${response.status} ${response.statusText}`);
	}

	return await response.json();
}

export async function getDiaryEntriesByMonth(month: MonthInput): Promise<DiaryEntryResult[]> {
	try {
		const response = await callDiaryEndpoint('/diary.DiaryService/GetDiaryEntriesByMonth', {
			month: createYM(month)
		});

		return response.entries.map(convertDiaryEntry);
	} catch (error) {
		// Handle specific gRPC errors
		if (error instanceof Error) {
			if (error.message.includes('401') || error.message.includes('UNAUTHENTICATED')) {
				throw new Error('Authentication required');
			} else if (error.message.includes('503') || error.message.includes('UNAVAILABLE')) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to get diary entries by month: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function updateDiaryEntry(request: UpdateDiaryRequest): Promise<DiaryEntryResult> {
	try {
		const response = await callDiaryEndpoint('/diary.DiaryService/UpdateDiaryEntry', {
			id: request.id,
			title: request.title,
			content: request.content,
			date: createYMD(request.date)
		});

		const entry = response.entry;
		if (!entry) {
			throw new Error('No diary entry returned');
		}

		return convertDiaryEntry(entry);
	} catch (error) {
		// Handle specific gRPC errors
		if (error instanceof Error) {
			if (error.message.includes('404') || error.message.includes('NOT_FOUND')) {
				throw new Error('Diary entry not found');
			} else if (error.message.includes('401') || error.message.includes('UNAUTHENTICATED')) {
				throw new Error('Authentication required');
			} else if (error.message.includes('503') || error.message.includes('UNAVAILABLE')) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to update diary entry: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function deleteDiaryEntry(id: string): Promise<boolean> {
	try {
		const response = await callDiaryEndpoint('/diary.DiaryService/DeleteDiaryEntry', { id });
		return response.success;
	} catch (error) {
		// Handle specific gRPC errors
		if (error instanceof Error) {
			if (error.message.includes('404') || error.message.includes('NOT_FOUND')) {
				throw new Error('Diary entry not found');
			} else if (error.message.includes('401') || error.message.includes('UNAUTHENTICATED')) {
				throw new Error('Authentication required');
			} else if (error.message.includes('503') || error.message.includes('UNAVAILABLE')) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to delete diary entry: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function searchDiaryEntries(request: SearchRequest): Promise<{ keyword: string; entries: DiaryEntryResult[] }> {
	try {
		const response = await callDiaryEndpoint('/diary.DiaryService/SearchDiaryEntries', {
			userID: request.userId,
			keyword: request.keyword
		});

		return {
			keyword: response.searchedKeyword,
			entries: response.entries.map(convertDiaryEntry)
		};
	} catch (error) {
		// Handle specific gRPC errors
		if (error instanceof Error) {
			if (error.message.includes('401') || error.message.includes('UNAUTHENTICATED')) {
				throw new Error('Authentication required');
			} else if (error.message.includes('503') || error.message.includes('UNAVAILABLE')) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to search diary entries: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}