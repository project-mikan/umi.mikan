import { writable } from "svelte/store";
import { browser } from "$app/environment";

// word-break: auto-phraseが有効かどうかを管理するストア
function createAutoPhraseStore() {
	// localStorageから初期値を読み込む
	const getInitialValue = (): boolean => {
		if (!browser) return true; // デフォルトで有効
		try {
			const stored = localStorage.getItem("auto-phrase-enabled");
			// 未設定の場合はtrueをデフォルトとする
			return stored === null ? true : stored === "true";
		} catch {
			return true;
		}
	};

	const { subscribe, set, update } = writable<boolean>(getInitialValue());

	return {
		subscribe,
		toggle: () => {
			update((value) => {
				const newValue = !value;
				// localStorageに保存
				if (browser) {
					try {
						localStorage.setItem("auto-phrase-enabled", String(newValue));
					} catch {
						// localStorage保存失敗時は無視
					}
				}
				return newValue;
			});
		},
		set: (value: boolean) => {
			set(value);
			// localStorageに保存
			if (browser) {
				try {
					localStorage.setItem("auto-phrase-enabled", String(value));
				} catch {
					// localStorage保存失敗時は無視
				}
			}
		},
	};
}

export const autoPhraseEnabled = createAutoPhraseStore();
