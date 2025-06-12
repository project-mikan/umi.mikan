// Original file: proto/diary/diary.proto

import type {
	DiaryEntry as _diary_DiaryEntry,
	DiaryEntry__Output as _diary_DiaryEntry__Output,
} from "../diary/DiaryEntry";

export interface SearchDiaryEntriesResponse {
	searched_keyword?: string;
	entries?: _diary_DiaryEntry[];
}

export interface SearchDiaryEntriesResponse__Output {
	searched_keyword: string;
	entries: _diary_DiaryEntry__Output[];
}
