/**
 * error-utilsのテスト
 */
import { describe, expect, it } from "vitest";
import { Code, ConnectError } from "@connectrpc/connect";
import { isInvalidRefreshTokenError } from "./error-utils";

describe("error-utils", () => {
  describe("isInvalidRefreshTokenError", () => {
    it("正常系: リフレッシュトークン自体が無効なエラーコードのみtrueを返し、一時的な障害はfalseを返す", () => {
      const testCases: {
        name: string;
        error: unknown;
        expected: boolean;
      }[] = [
        {
          name: "InvalidArgument: トークンが期限切れ・改竄されている",
          error: new ConnectError("validation error", Code.InvalidArgument),
          expected: true,
        },
        {
          name: "Unauthenticated: ユーザーが存在しない",
          error: new ConnectError("user not found", Code.Unauthenticated),
          expected: true,
        },
        {
          name: "Internal: バックエンドの内部エラー",
          error: new ConnectError("failed to get user by ID", Code.Internal),
          expected: false,
        },
        {
          name: "Unavailable: バックエンドに接続できない",
          error: new ConnectError("connection refused", Code.Unavailable),
          expected: false,
        },
        {
          name: "DeadlineExceeded: リクエストがタイムアウトした",
          error: new ConnectError("deadline exceeded", Code.DeadlineExceeded),
          expected: false,
        },
        {
          name: "ConnectError以外の例外",
          error: new Error("network error"),
          expected: false,
        },
        {
          name: "errorがundefined",
          error: undefined,
          expected: false,
        },
      ];

      for (const testCase of testCases) {
        expect(isInvalidRefreshTokenError(testCase.error), testCase.name).toBe(
          testCase.expected,
        );
      }
    });
  });
});
