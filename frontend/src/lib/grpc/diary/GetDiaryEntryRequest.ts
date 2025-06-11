// Original file: proto/diary/diary.proto

import type { YMD as _diary_YMD, YMD__Output as _diary_YMD__Output } from '../diary/YMD';

export interface GetDiaryEntryRequest {
  'date'?: (_diary_YMD | null);
}

export interface GetDiaryEntryRequest__Output {
  'date': (_diary_YMD__Output | null);
}
