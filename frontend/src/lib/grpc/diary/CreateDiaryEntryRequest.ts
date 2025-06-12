// Original file: proto/diary/diary.proto

import type {
	YMD as _diary_YMD,
	YMD__Output as _diary_YMD__Output,
} from "../diary/YMD";

export interface CreateDiaryEntryRequest {
	content?: string;
	date?: _diary_YMD | null;
}

export interface CreateDiaryEntryRequest__Output {
	content: string;
	date: _diary_YMD__Output | null;
}
