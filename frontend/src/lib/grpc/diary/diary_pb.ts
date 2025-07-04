// @generated by protoc-gen-es v2.5.2 with parameter "target=ts"
// @generated from file diary/diary.proto (package diary, syntax proto3)
/* eslint-disable */

import type { Message } from "@bufbuild/protobuf";
import type {
	GenFile,
	GenMessage,
	GenService,
} from "@bufbuild/protobuf/codegenv2";
import {
	fileDesc,
	messageDesc,
	serviceDesc,
} from "@bufbuild/protobuf/codegenv2";

/**
 * Describes the file diary/diary.proto.
 */
export const file_diary_diary: GenFile =
	/*@__PURE__*/
	fileDesc(
		"ChFkaWFyeS9kaWFyeS5wcm90bxIFZGlhcnkiLwoDWU1EEgwKBHllYXIYASABKA0SDQoFbW9udGgYAiABKA0SCwoDZGF5GAMgASgNIiEKAllNEgwKBHllYXIYASABKA0SDQoFbW9udGgYAiABKA0iQwoKRGlhcnlFbnRyeRIKCgJpZBgBIAEoCRIYCgRkYXRlGAIgASgLMgouZGlhcnkuWU1EEg8KB2NvbnRlbnQYAyABKAkiRAoXQ3JlYXRlRGlhcnlFbnRyeVJlcXVlc3QSDwoHY29udGVudBgBIAEoCRIYCgRkYXRlGAIgASgLMgouZGlhcnkuWU1EIjwKGENyZWF0ZURpYXJ5RW50cnlSZXNwb25zZRIgCgVlbnRyeRgBIAEoCzIRLmRpYXJ5LkRpYXJ5RW50cnkiMAoUR2V0RGlhcnlFbnRyeVJlcXVlc3QSGAoEZGF0ZRgBIAEoCzIKLmRpYXJ5LllNRCIzChZHZXREaWFyeUVudHJpZXNSZXF1ZXN0EhkKBWRhdGVzGAEgAygLMgouZGlhcnkuWU1EIjkKHUdldERpYXJ5RW50cmllc0J5TW9udGhSZXF1ZXN0EhgKBW1vbnRoGAEgASgLMgkuZGlhcnkuWU0iLAoZU2VhcmNoRGlhcnlFbnRyaWVzUmVxdWVzdBIPCgdrZXl3b3JkGAEgASgJIloKGlNlYXJjaERpYXJ5RW50cmllc1Jlc3BvbnNlEhgKEHNlYXJjaGVkX2tleXdvcmQYASABKAkSIgoHZW50cmllcxgCIAMoCzIRLmRpYXJ5LkRpYXJ5RW50cnkiPQoXR2V0RGlhcnlFbnRyaWVzUmVzcG9uc2USIgoHZW50cmllcxgBIAMoCzIRLmRpYXJ5LkRpYXJ5RW50cnkiRAoeR2V0RGlhcnlFbnRyaWVzQnlNb250aFJlc3BvbnNlEiIKB2VudHJpZXMYASADKAsyES5kaWFyeS5EaWFyeUVudHJ5IjkKFUdldERpYXJ5RW50cnlSZXNwb25zZRIgCgVlbnRyeRgBIAEoCzIRLmRpYXJ5LkRpYXJ5RW50cnkiXwoXVXBkYXRlRGlhcnlFbnRyeVJlcXVlc3QSCgoCaWQYASABKAkSDQoFdGl0bGUYAiABKAkSDwoHY29udGVudBgDIAEoCRIYCgRkYXRlGAQgASgLMgouZGlhcnkuWU1EIjwKGFVwZGF0ZURpYXJ5RW50cnlSZXNwb25zZRIgCgVlbnRyeRgBIAEoCzIRLmRpYXJ5LkRpYXJ5RW50cnkiJQoXRGVsZXRlRGlhcnlFbnRyeVJlcXVlc3QSCgoCaWQYASABKAkiKwoYRGVsZXRlRGlhcnlFbnRyeVJlc3BvbnNlEg8KB3N1Y2Nlc3MYASABKAgy7QQKDERpYXJ5U2VydmljZRJTChBDcmVhdGVEaWFyeUVudHJ5Eh4uZGlhcnkuQ3JlYXRlRGlhcnlFbnRyeVJlcXVlc3QaHy5kaWFyeS5DcmVhdGVEaWFyeUVudHJ5UmVzcG9uc2USUwoQVXBkYXRlRGlhcnlFbnRyeRIeLmRpYXJ5LlVwZGF0ZURpYXJ5RW50cnlSZXF1ZXN0Gh8uZGlhcnkuVXBkYXRlRGlhcnlFbnRyeVJlc3BvbnNlElMKEERlbGV0ZURpYXJ5RW50cnkSHi5kaWFyeS5EZWxldGVEaWFyeUVudHJ5UmVxdWVzdBofLmRpYXJ5LkRlbGV0ZURpYXJ5RW50cnlSZXNwb25zZRJKCg1HZXREaWFyeUVudHJ5EhsuZGlhcnkuR2V0RGlhcnlFbnRyeVJlcXVlc3QaHC5kaWFyeS5HZXREaWFyeUVudHJ5UmVzcG9uc2USUAoPR2V0RGlhcnlFbnRyaWVzEh0uZGlhcnkuR2V0RGlhcnlFbnRyaWVzUmVxdWVzdBoeLmRpYXJ5LkdldERpYXJ5RW50cmllc1Jlc3BvbnNlEmUKFkdldERpYXJ5RW50cmllc0J5TW9udGgSJC5kaWFyeS5HZXREaWFyeUVudHJpZXNCeU1vbnRoUmVxdWVzdBolLmRpYXJ5LkdldERpYXJ5RW50cmllc0J5TW9udGhSZXNwb25zZRJZChJTZWFyY2hEaWFyeUVudHJpZXMSIC5kaWFyeS5TZWFyY2hEaWFyeUVudHJpZXNSZXF1ZXN0GiEuZGlhcnkuU2VhcmNoRGlhcnlFbnRyaWVzUmVzcG9uc2VCQFo+Z2l0aHViLmNvbS9wcm9qZWN0LW1pa2FuL3VtaS5taWthbi9iYWNrZW5kL2luZnJhc3RydWN0dXJlL2dycGNiBnByb3RvMw",
	);

/**
 * @generated from message diary.YMD
 */
export type YMD = Message<"diary.YMD"> & {
	/**
	 * @generated from field: uint32 year = 1;
	 */
	year: number;

	/**
	 * @generated from field: uint32 month = 2;
	 */
	month: number;

	/**
	 * @generated from field: uint32 day = 3;
	 */
	day: number;
};

/**
 * Describes the message diary.YMD.
 * Use `create(YMDSchema)` to create a new message.
 */
export const YMDSchema: GenMessage<YMD> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 0);

/**
 * @generated from message diary.YM
 */
export type YM = Message<"diary.YM"> & {
	/**
	 * @generated from field: uint32 year = 1;
	 */
	year: number;

	/**
	 * @generated from field: uint32 month = 2;
	 */
	month: number;
};

/**
 * Describes the message diary.YM.
 * Use `create(YMSchema)` to create a new message.
 */
export const YMSchema: GenMessage<YM> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 1);

/**
 * 日記エントリのメッセージ
 *
 * @generated from message diary.DiaryEntry
 */
export type DiaryEntry = Message<"diary.DiaryEntry"> & {
	/**
	 * 日記ID
	 *
	 * @generated from field: string id = 1;
	 */
	id: string;

	/**
	 * 日付
	 *
	 * @generated from field: diary.YMD date = 2;
	 */
	date?: YMD;

	/**
	 * 内容
	 *
	 * @generated from field: string content = 3;
	 */
	content: string;
};

/**
 * Describes the message diary.DiaryEntry.
 * Use `create(DiaryEntrySchema)` to create a new message.
 */
export const DiaryEntrySchema: GenMessage<DiaryEntry> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 2);

/**
 * 新しい日記エントリを作成するためのリクエスト
 *
 * @generated from message diary.CreateDiaryEntryRequest
 */
export type CreateDiaryEntryRequest =
	Message<"diary.CreateDiaryEntryRequest"> & {
		/**
		 * @generated from field: string content = 1;
		 */
		content: string;

		/**
		 * @generated from field: diary.YMD date = 2;
		 */
		date?: YMD;
	};

/**
 * Describes the message diary.CreateDiaryEntryRequest.
 * Use `create(CreateDiaryEntryRequestSchema)` to create a new message.
 */
export const CreateDiaryEntryRequestSchema: GenMessage<CreateDiaryEntryRequest> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 3);

/**
 * 日記エントリを作成した結果を返すレスポンス
 *
 * @generated from message diary.CreateDiaryEntryResponse
 */
export type CreateDiaryEntryResponse =
	Message<"diary.CreateDiaryEntryResponse"> & {
		/**
		 * @generated from field: diary.DiaryEntry entry = 1;
		 */
		entry?: DiaryEntry;
	};

/**
 * Describes the message diary.CreateDiaryEntryResponse.
 * Use `create(CreateDiaryEntryResponseSchema)` to create a new message.
 */
export const CreateDiaryEntryResponseSchema: GenMessage<CreateDiaryEntryResponse> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 4);

/**
 * 特定の日記エントリを取得するためのリクエスト
 *
 * @generated from message diary.GetDiaryEntryRequest
 */
export type GetDiaryEntryRequest = Message<"diary.GetDiaryEntryRequest"> & {
	/**
	 * 日付を指定して取得
	 *
	 * @generated from field: diary.YMD date = 1;
	 */
	date?: YMD;
};

/**
 * Describes the message diary.GetDiaryEntryRequest.
 * Use `create(GetDiaryEntryRequestSchema)` to create a new message.
 */
export const GetDiaryEntryRequestSchema: GenMessage<GetDiaryEntryRequest> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 5);

/**
 * 複数日記エントリを取得するためのリクエスト (e.g., by range or count)
 *
 * @generated from message diary.GetDiaryEntriesRequest
 */
export type GetDiaryEntriesRequest = Message<"diary.GetDiaryEntriesRequest"> & {
	/**
	 * 取得したい日付の配列
	 *
	 * @generated from field: repeated diary.YMD dates = 1;
	 */
	dates: YMD[];
};

/**
 * Describes the message diary.GetDiaryEntriesRequest.
 * Use `create(GetDiaryEntriesRequestSchema)` to create a new message.
 */
export const GetDiaryEntriesRequestSchema: GenMessage<GetDiaryEntriesRequest> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 6);

/**
 * 月ごとに日記エントリを取得するためのリクエスト
 *
 * @generated from message diary.GetDiaryEntriesByMonthRequest
 */
export type GetDiaryEntriesByMonthRequest =
	Message<"diary.GetDiaryEntriesByMonthRequest"> & {
		/**
		 * 年月を指定
		 *
		 * @generated from field: diary.YM month = 1;
		 */
		month?: YM;
	};

/**
 * Describes the message diary.GetDiaryEntriesByMonthRequest.
 * Use `create(GetDiaryEntriesByMonthRequestSchema)` to create a new message.
 */
export const GetDiaryEntriesByMonthRequestSchema: GenMessage<GetDiaryEntriesByMonthRequest> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 7);

/**
 * @generated from message diary.SearchDiaryEntriesRequest
 */
export type SearchDiaryEntriesRequest =
	Message<"diary.SearchDiaryEntriesRequest"> & {
		/**
		 * @generated from field: string keyword = 1;
		 */
		keyword: string;
	};

/**
 * Describes the message diary.SearchDiaryEntriesRequest.
 * Use `create(SearchDiaryEntriesRequestSchema)` to create a new message.
 */
export const SearchDiaryEntriesRequestSchema: GenMessage<SearchDiaryEntriesRequest> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 8);

/**
 * @generated from message diary.SearchDiaryEntriesResponse
 */
export type SearchDiaryEntriesResponse =
	Message<"diary.SearchDiaryEntriesResponse"> & {
		/**
		 * 実際に検索した単語
		 *
		 * @generated from field: string searched_keyword = 1;
		 */
		searchedKeyword: string;

		/**
		 * @generated from field: repeated diary.DiaryEntry entries = 2;
		 */
		entries: DiaryEntry[];
	};

/**
 * Describes the message diary.SearchDiaryEntriesResponse.
 * Use `create(SearchDiaryEntriesResponseSchema)` to create a new message.
 */
export const SearchDiaryEntriesResponseSchema: GenMessage<SearchDiaryEntriesResponse> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 9);

/**
 * @generated from message diary.GetDiaryEntriesResponse
 */
export type GetDiaryEntriesResponse =
	Message<"diary.GetDiaryEntriesResponse"> & {
		/**
		 * @generated from field: repeated diary.DiaryEntry entries = 1;
		 */
		entries: DiaryEntry[];
	};

/**
 * Describes the message diary.GetDiaryEntriesResponse.
 * Use `create(GetDiaryEntriesResponseSchema)` to create a new message.
 */
export const GetDiaryEntriesResponseSchema: GenMessage<GetDiaryEntriesResponse> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 10);

/**
 * @generated from message diary.GetDiaryEntriesByMonthResponse
 */
export type GetDiaryEntriesByMonthResponse =
	Message<"diary.GetDiaryEntriesByMonthResponse"> & {
		/**
		 * @generated from field: repeated diary.DiaryEntry entries = 1;
		 */
		entries: DiaryEntry[];
	};

/**
 * Describes the message diary.GetDiaryEntriesByMonthResponse.
 * Use `create(GetDiaryEntriesByMonthResponseSchema)` to create a new message.
 */
export const GetDiaryEntriesByMonthResponseSchema: GenMessage<GetDiaryEntriesByMonthResponse> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 11);

/**
 * 日記エントリを取得した結果を返すレスポンス
 *
 * @generated from message diary.GetDiaryEntryResponse
 */
export type GetDiaryEntryResponse = Message<"diary.GetDiaryEntryResponse"> & {
	/**
	 * @generated from field: diary.DiaryEntry entry = 1;
	 */
	entry?: DiaryEntry;
};

/**
 * Describes the message diary.GetDiaryEntryResponse.
 * Use `create(GetDiaryEntryResponseSchema)` to create a new message.
 */
export const GetDiaryEntryResponseSchema: GenMessage<GetDiaryEntryResponse> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 12);

/**
 * 日記エントリを更新するためのリクエスト
 *
 * @generated from message diary.UpdateDiaryEntryRequest
 */
export type UpdateDiaryEntryRequest =
	Message<"diary.UpdateDiaryEntryRequest"> & {
		/**
		 * @generated from field: string id = 1;
		 */
		id: string;

		/**
		 * @generated from field: string title = 2;
		 */
		title: string;

		/**
		 * @generated from field: string content = 3;
		 */
		content: string;

		/**
		 * @generated from field: diary.YMD date = 4;
		 */
		date?: YMD;
	};

/**
 * Describes the message diary.UpdateDiaryEntryRequest.
 * Use `create(UpdateDiaryEntryRequestSchema)` to create a new message.
 */
export const UpdateDiaryEntryRequestSchema: GenMessage<UpdateDiaryEntryRequest> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 13);

/**
 * 更新された日記エントリを返すレスポンス
 *
 * @generated from message diary.UpdateDiaryEntryResponse
 */
export type UpdateDiaryEntryResponse =
	Message<"diary.UpdateDiaryEntryResponse"> & {
		/**
		 * @generated from field: diary.DiaryEntry entry = 1;
		 */
		entry?: DiaryEntry;
	};

/**
 * Describes the message diary.UpdateDiaryEntryResponse.
 * Use `create(UpdateDiaryEntryResponseSchema)` to create a new message.
 */
export const UpdateDiaryEntryResponseSchema: GenMessage<UpdateDiaryEntryResponse> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 14);

/**
 * 日記エントリを削除するためのリクエスト
 *
 * @generated from message diary.DeleteDiaryEntryRequest
 */
export type DeleteDiaryEntryRequest =
	Message<"diary.DeleteDiaryEntryRequest"> & {
		/**
		 * @generated from field: string id = 1;
		 */
		id: string;
	};

/**
 * Describes the message diary.DeleteDiaryEntryRequest.
 * Use `create(DeleteDiaryEntryRequestSchema)` to create a new message.
 */
export const DeleteDiaryEntryRequestSchema: GenMessage<DeleteDiaryEntryRequest> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 15);

/**
 * 削除操作の結果を返すレスポンス
 *
 * @generated from message diary.DeleteDiaryEntryResponse
 */
export type DeleteDiaryEntryResponse =
	Message<"diary.DeleteDiaryEntryResponse"> & {
		/**
		 * @generated from field: bool success = 1;
		 */
		success: boolean;
	};

/**
 * Describes the message diary.DeleteDiaryEntryResponse.
 * Use `create(DeleteDiaryEntryResponseSchema)` to create a new message.
 */
export const DeleteDiaryEntryResponseSchema: GenMessage<DeleteDiaryEntryResponse> =
	/*@__PURE__*/
	messageDesc(file_diary_diary, 16);

/**
 * @generated from service diary.DiaryService
 */
export const DiaryService: GenService<{
	/**
	 * 作成
	 *
	 * @generated from rpc diary.DiaryService.CreateDiaryEntry
	 */
	createDiaryEntry: {
		methodKind: "unary";
		input: typeof CreateDiaryEntryRequestSchema;
		output: typeof CreateDiaryEntryResponseSchema;
	};
	/**
	 * 更新
	 *
	 * @generated from rpc diary.DiaryService.UpdateDiaryEntry
	 */
	updateDiaryEntry: {
		methodKind: "unary";
		input: typeof UpdateDiaryEntryRequestSchema;
		output: typeof UpdateDiaryEntryResponseSchema;
	};
	/**
	 * 削除
	 *
	 * @generated from rpc diary.DiaryService.DeleteDiaryEntry
	 */
	deleteDiaryEntry: {
		methodKind: "unary";
		input: typeof DeleteDiaryEntryRequestSchema;
		output: typeof DeleteDiaryEntryResponseSchema;
	};
	/**
	 * 日付指定で単体取得
	 *
	 * @generated from rpc diary.DiaryService.GetDiaryEntry
	 */
	getDiaryEntry: {
		methodKind: "unary";
		input: typeof GetDiaryEntryRequestSchema;
		output: typeof GetDiaryEntryResponseSchema;
	};
	/**
	 * 日付指定で複数取得(ホームでの表示などで直近3日とかほしいケースや過去数年分ほしいケースに対応)
	 *
	 * @generated from rpc diary.DiaryService.GetDiaryEntries
	 */
	getDiaryEntries: {
		methodKind: "unary";
		input: typeof GetDiaryEntriesRequestSchema;
		output: typeof GetDiaryEntriesResponseSchema;
	};
	/**
	 * 月ごとに取得
	 *
	 * @generated from rpc diary.DiaryService.GetDiaryEntriesByMonth
	 */
	getDiaryEntriesByMonth: {
		methodKind: "unary";
		input: typeof GetDiaryEntriesByMonthRequestSchema;
		output: typeof GetDiaryEntriesByMonthResponseSchema;
	};
	/**
	 * 検索
	 *
	 * @generated from rpc diary.DiaryService.SearchDiaryEntries
	 */
	searchDiaryEntries: {
		methodKind: "unary";
		input: typeof SearchDiaryEntriesRequestSchema;
		output: typeof SearchDiaryEntriesResponseSchema;
	};
}> = /*@__PURE__*/ serviceDesc(file_diary_diary, 0);
