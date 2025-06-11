// Original file: proto/diary/diary.proto

import type { YMD as _diary_YMD, YMD__Output as _diary_YMD__Output } from '../diary/YMD';

export interface GetDiaryEntriesRequest {
  'dates'?: (_diary_YMD)[];
}

export interface GetDiaryEntriesRequest__Output {
  'dates': (_diary_YMD__Output)[];
}
