import { writable } from "svelte/store";
import { browser } from "$app/environment";

// BudouXが有効かどうかを管理するストア
function createBudouXStore() {
	// localStorageから初期値を読み込む
	const getInitialValue = (): boolean => {
		if (!browser) return false;
		try {
			const stored = localStorage.getItem("budoux-enabled");
			return stored === "true";
		} catch {
			return false;
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
						localStorage.setItem("budoux-enabled", String(newValue));
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
					localStorage.setItem("budoux-enabled", String(value));
				} catch {
					// localStorage保存失敗時は無視
				}
			}
		},
	};
}

export const budouxEnabled = createBudouXStore();
