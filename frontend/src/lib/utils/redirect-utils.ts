/**
 * ログイン後の戻り先として渡された値を、同一サイト内のパスに限って返す。
 * "//evil.example" や "/\evil.example" のようなプロトコル相対URL、絶対URLは
 * オープンリダイレクトになるため拒否し、トップページ（"/"）にフォールバックする。
 */
export function getSafeRedirectPath(value: string | null): string {
  if (!value?.startsWith("/")) {
    return "/";
  }
  if (value.startsWith("//") || value.startsWith("/\\")) {
    return "/";
  }
  return value;
}
