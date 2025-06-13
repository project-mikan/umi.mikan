// gRPC APIを使用したバックエンドとの通信（Diary Service）
import { createDiaryClient, promisifyGrpcCall } from './grpc-client';
import { 
	CreateDiaryEntryRequest,
	CreateDiaryEntryResponse,
	GetDiaryEntryRequest,
	GetDiaryEntryResponse,
	GetDiaryEntriesRequest,
	GetDiaryEntriesResponse,
	GetDiaryEntriesByMonthRequest,
	GetDiaryEntriesByMonthResponse,
	UpdateDiaryEntryRequest,
	UpdateDiaryEntryResponse,
	DeleteDiaryEntryRequest,
	DeleteDiaryEntryResponse,
	SearchDiaryEntriesRequest,
	SearchDiaryEntriesResponse,
	DiaryEntry,
	YMD,
	YM
} from '../grpc/diary/diary_pb';
import * as grpc from '@grpc/grpc-js';

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
	const ymd = new YMD();
	ymd.setYear(date.year);
	ymd.setMonth(date.month);
	ymd.setDay(date.day);
	return ymd;
}

// Helper function to create YM message
function createYM(month: MonthInput): YM {
	const ym = new YM();
	ym.setYear(month.year);
	ym.setMonth(month.month);
	return ym;
}

// Helper function to convert DiaryEntry to result
function convertDiaryEntry(entry: DiaryEntry): DiaryEntryResult {
	const date = entry.getDate();
	return {
		id: entry.getId(),
		date: {
			year: date?.getYear() || 0,
			month: date?.getMonth() || 0,
			day: date?.getDay() || 0
		},
		content: entry.getContent()
	};
}

// API functions
export async function createDiaryEntry(request: CreateDiaryRequest): Promise<DiaryEntryResult> {
	try {
		const client = createDiaryClient();
		
		const grpcRequest = new CreateDiaryEntryRequest();
		grpcRequest.setContent(request.content);
		grpcRequest.setDate(createYMD(request.date));

		const response = await promisifyGrpcCall<CreateDiaryEntryRequest, CreateDiaryEntryResponse>(
			client,
			'createDiaryEntry',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		const entry = response.getEntry();
		if (!entry) {
			throw new Error('No diary entry returned');
		}

		return convertDiaryEntry(entry);
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.ALREADY_EXISTS) {
				throw new Error('Diary entry for this date already exists');
			} else if (grpcError.code === grpc.status.UNAUTHENTICATED) {
				throw new Error('Authentication required');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to create diary entry: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function getDiaryEntry(date: DateInput): Promise<DiaryEntryResult | null> {
	try {
		const client = createDiaryClient();
		
		const grpcRequest = new GetDiaryEntryRequest();
		grpcRequest.setDate(createYMD(date));

		const response = await promisifyGrpcCall<GetDiaryEntryRequest, GetDiaryEntryResponse>(
			client,
			'getDiaryEntry',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		const entry = response.getEntry();
		if (!entry) {
			return null;
		}

		return convertDiaryEntry(entry);
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.NOT_FOUND) {
				return null;
			} else if (grpcError.code === grpc.status.UNAUTHENTICATED) {
				throw new Error('Authentication required');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to get diary entry: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function getDiaryEntries(dates: DateInput[]): Promise<DiaryEntryResult[]> {
	try {
		const client = createDiaryClient();
		
		const grpcRequest = new GetDiaryEntriesRequest();
		const ymdDates = dates.map(createYMD);
		grpcRequest.setDatesList(ymdDates);

		const response = await promisifyGrpcCall<GetDiaryEntriesRequest, GetDiaryEntriesResponse>(
			client,
			'getDiaryEntries',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		return response.getEntriesList().map(convertDiaryEntry);
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.UNAUTHENTICATED) {
				throw new Error('Authentication required');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to get diary entries: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function getDiaryEntriesByMonth(month: MonthInput): Promise<DiaryEntryResult[]> {
	try {
		const client = createDiaryClient();
		
		const grpcRequest = new GetDiaryEntriesByMonthRequest();
		grpcRequest.setMonth(createYM(month));

		const response = await promisifyGrpcCall<GetDiaryEntriesByMonthRequest, GetDiaryEntriesByMonthResponse>(
			client,
			'getDiaryEntriesByMonth',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		return response.getEntriesList().map(convertDiaryEntry);
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.UNAUTHENTICATED) {
				throw new Error('Authentication required');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to get diary entries by month: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function updateDiaryEntry(request: UpdateDiaryRequest): Promise<DiaryEntryResult> {
	try {
		const client = createDiaryClient();
		
		const grpcRequest = new UpdateDiaryEntryRequest();
		grpcRequest.setId(request.id);
		grpcRequest.setTitle(request.title);
		grpcRequest.setContent(request.content);
		grpcRequest.setDate(createYMD(request.date));

		const response = await promisifyGrpcCall<UpdateDiaryEntryRequest, UpdateDiaryEntryResponse>(
			client,
			'updateDiaryEntry',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		const entry = response.getEntry();
		if (!entry) {
			throw new Error('No diary entry returned');
		}

		return convertDiaryEntry(entry);
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.NOT_FOUND) {
				throw new Error('Diary entry not found');
			} else if (grpcError.code === grpc.status.UNAUTHENTICATED) {
				throw new Error('Authentication required');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to update diary entry: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function deleteDiaryEntry(id: string): Promise<boolean> {
	try {
		const client = createDiaryClient();
		
		const grpcRequest = new DeleteDiaryEntryRequest();
		grpcRequest.setId(id);

		const response = await promisifyGrpcCall<DeleteDiaryEntryRequest, DeleteDiaryEntryResponse>(
			client,
			'deleteDiaryEntry',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		return response.getSuccess();
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.NOT_FOUND) {
				throw new Error('Diary entry not found');
			} else if (grpcError.code === grpc.status.UNAUTHENTICATED) {
				throw new Error('Authentication required');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to delete diary entry: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function searchDiaryEntries(request: SearchRequest): Promise<{ keyword: string; entries: DiaryEntryResult[] }> {
	try {
		const client = createDiaryClient();
		
		const grpcRequest = new SearchDiaryEntriesRequest();
		grpcRequest.setUserid(request.userId);
		grpcRequest.setKeyword(request.keyword);

		const response = await promisifyGrpcCall<SearchDiaryEntriesRequest, SearchDiaryEntriesResponse>(
			client,
			'searchDiaryEntries',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		return {
			keyword: response.getSearchedKeyword(),
			entries: response.getEntriesList().map(convertDiaryEntry)
		};
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.UNAUTHENTICATED) {
				throw new Error('Authentication required');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Failed to search diary entries: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}