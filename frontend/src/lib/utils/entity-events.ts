/**
 * エンティティ更新イベントを発火
 * エンティティが作成、更新、削除された時に呼び出す
 */
export function notifyEntityUpdated() {
	if (typeof window !== "undefined") {
		const event = new CustomEvent("entityUpdated");
		window.dispatchEvent(event);
	}
}
