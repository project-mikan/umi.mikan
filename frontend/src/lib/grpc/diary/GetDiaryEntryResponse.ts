// Original file: proto/diary/diary.proto

import type {
	DiaryEntry as _diary_DiaryEntry,
	DiaryEntry__Output as _diary_DiaryEntry__Output,
} from "../diary/DiaryEntry";

export interface GetDiaryEntryResponse {
	entry?: _diary_DiaryEntry | null;
}

export interface GetDiaryEntryResponse__Output {
	entry: _diary_DiaryEntry__Output | null;
}
