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
		// より厳密な正規表現でセキュリティを向上
		const match = message.match(/try again in (\d+(?:[smh])+)(?:\s|$)/);
		if (match && match[1]) {
			return match[1];
		}
	}
	return null;
}

/**
 * エラーメッセージを日本語に変換（セキュリティを考慮したメッセージ）
 */
export function translateErrorMessage(error: unknown): string {
	if (error instanceof ConnectError) {
		switch (error.code) {
			case Code.ResourceExhausted:
				// レート制限の場合のみリセット時間を表示
				return "アクセス制限中です。しばらく時間をおいて再試行してください。";
			case Code.Unauthenticated:
				// 認証失敗の詳細は明かさない
				return "ログインに失敗しました。入力内容を確認してください。";
			case Code.NotFound:
				// 具体的にユーザーが見つからないことは明かさない
				return "ログインに失敗しました。入力内容を確認してください。";
			case Code.AlreadyExists:
				return "このメールアドレスは既に登録されています。";
			case Code.InvalidArgument:
				return "入力内容に不備があります。確認してください。";
			case Code.Internal:
			case Code.Unavailable:
				return "サービスが一時的に利用できません。しばらく時間をおいて再試行してください。";
			default:
				return "エラーが発生しました。しばらく時間をおいて再試行してください。";
		}
	}

	// セキュリティのため、詳細なエラーメッセージは表示しない
	return "エラーが発生しました。しばらく時間をおいて再試行してください。";
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