import { ConnectError } from "@connectrpc/connect";
import { Code } from "@connectrpc/connect";

/**
 * レート制限エラーかどうかを判定
 */
export function isRateLimitError(error: unknown): boolean {
	if (error instanceof ConnectError) {
		return error.code === Code.ResourceExhausted;
	}
	return false;
}

/**
 * レート制限エラーからリセット時間を抽出（可能であれば）
 */
export function extractRateLimitResetTime(error: unknown): string | null {
	if (error instanceof ConnectError && error.code === Code.ResourceExhausted) {
		const message = error.message;
		// "try again in 14m32s" のような形式からリセット時間を抽出
		const match = message.match(/try again in (\d+[smh]+)/);
		if (match) {
			return match[1];
		}
	}
	return null;
}

/**
 * エラーメッセージを日本語に変換
 */
export function translateErrorMessage(error: unknown): string {
	if (error instanceof ConnectError) {
		switch (error.code) {
			case Code.ResourceExhausted:
				const resetTime = extractRateLimitResetTime(error);
				if (resetTime) {
					return `ログイン試行回数の上限に達しました。${resetTime}後に再試行してください。`;
				}
				return "ログイン試行回数の上限に達しました。しばらく時間をおいて再試行してください。";
			case Code.Unauthenticated:
				return "認証に失敗しました。メールアドレスまたはパスワードが正しくありません。";
			case Code.NotFound:
				return "ユーザーが見つかりません。";
			case Code.AlreadyExists:
				return "このメールアドレスは既に登録されています。";
			case Code.InvalidArgument:
				return "入力データが正しくありません。";
			case Code.Internal:
				return "内部エラーが発生しました。しばらく時間をおいて再試行してください。";
			case Code.Unavailable:
				return "サービスが一時的に利用できません。しばらく時間をおいて再試行してください。";
			default:
				return "エラーが発生しました。しばらく時間をおいて再試行してください。";
		}
	}

	if (error instanceof Error) {
		return error.message;
	}

	return "予期しないエラーが発生しました。";
}

/**
 * エラーの重要度を判定
 */
export function getErrorSeverity(error: unknown): "info" | "warning" | "error" {
	if (error instanceof ConnectError) {
		switch (error.code) {
			case Code.ResourceExhausted:
			case Code.Unauthenticated:
			case Code.NotFound:
			case Code.AlreadyExists:
			case Code.InvalidArgument:
				return "warning";
			case Code.Internal:
			case Code.Unavailable:
				return "error";
			default:
				return "warning";
		}
	}
	return "error";
}