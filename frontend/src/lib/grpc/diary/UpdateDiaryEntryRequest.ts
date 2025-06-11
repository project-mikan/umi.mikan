// Original file: proto/diary/diary.proto

import type { YMD as _diary_YMD, YMD__Output as _diary_YMD__Output } from '../diary/YMD';

export interface UpdateDiaryEntryRequest {
  'id'?: (string);
  'title'?: (string);
  'content'?: (string);
  'date'?: (_diary_YMD | null);
}

export interface UpdateDiaryEntryRequest__Output {
  'id': (string);
  'title': (string);
  'content': (string);
  'date': (_diary_YMD__Output | null);
}
