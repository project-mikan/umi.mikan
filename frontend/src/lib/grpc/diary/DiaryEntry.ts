// Original file: proto/diary/diary.proto

import type { YMD as _diary_YMD, YMD__Output as _diary_YMD__Output } from '../diary/YMD';

export interface DiaryEntry {
  'id'?: (string);
  'date'?: (_diary_YMD | null);
  'content'?: (string);
}

export interface DiaryEntry__Output {
  'id': (string);
  'date': (_diary_YMD__Output | null);
  'content': (string);
}
