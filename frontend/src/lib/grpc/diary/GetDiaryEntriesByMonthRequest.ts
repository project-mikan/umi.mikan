// Original file: proto/diary/diary.proto

import type { YM as _diary_YM, YM__Output as _diary_YM__Output } from '../diary/YM';

export interface GetDiaryEntriesByMonthRequest {
  'month'?: (_diary_YM | null);
}

export interface GetDiaryEntriesByMonthRequest__Output {
  'month': (_diary_YM__Output | null);
}
