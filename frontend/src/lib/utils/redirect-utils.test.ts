import { describe, expect, it } from "vitest";
import { getSafeRedirectPath } from "./redirect-utils";

describe("getSafeRedirectPath", () => {
  it("同一サイト内のパス（クエリ付き）はそのまま返す", () => {
    expect(getSafeRedirectPath("/oauth/authorize?client_id=a&state=b")).toBe(
      "/oauth/authorize?client_id=a&state=b",
    );
  });

  it("未指定・空文字の場合はトップページを返す", () => {
    expect(getSafeRedirectPath(null)).toBe("/");
    expect(getSafeRedirectPath("")).toBe("/");
  });

  it("絶対URLはオープンリダイレクトになるためトップページを返す", () => {
    expect(getSafeRedirectPath("https://evil.example/")).toBe("/");
  });

  it("プロトコル相対URLはオープンリダイレクトになるためトップページを返す", () => {
    expect(getSafeRedirectPath("//evil.example/")).toBe("/");
    expect(getSafeRedirectPath("/\\evil.example/")).toBe("/");
  });
});
