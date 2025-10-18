import { writable } from "svelte/store";
import { browser } from "$app/environment";

export interface SummaryVisibilityState {
	daily: boolean;
	monthly: boolean;
	latestTrend: boolean;
}

const DEFAULT_STATE: SummaryVisibilityState = {
	daily: true,
	monthly: true,
	latestTrend: true,
};

function createSummaryVisibilityStore() {
	const { subscribe, set, update } =
		writable<SummaryVisibilityState>(DEFAULT_STATE);

	let isInitialized = false;

	function loadFromLocalStorage(): SummaryVisibilityState {
		if (!browser) return DEFAULT_STATE;

		try {
			const stored = localStorage.getItem("summary-visibility");
			if (stored) {
				return { ...DEFAULT_STATE, ...JSON.parse(stored) };
			}
		} catch (_error) {
			// Failed to load from localStorage, using default state
		}
		return DEFAULT_STATE;
	}

	function saveToLocalStorage(state: SummaryVisibilityState) {
		if (!browser) return;

		try {
			localStorage.setItem("summary-visibility", JSON.stringify(state));
		} catch (_error) {
			// Failed to save to localStorage
		}
	}

	function toggleDaily() {
		update((state) => {
			const newState = { ...state, daily: !state.daily };
			saveToLocalStorage(newState);
			return newState;
		});
	}

	function toggleMonthly() {
		update((state) => {
			const newState = { ...state, monthly: !state.monthly };
			saveToLocalStorage(newState);
			return newState;
		});
	}

	function toggleLatestTrend() {
		update((state) => {
			const newState = { ...state, latestTrend: !state.latestTrend };
			saveToLocalStorage(newState);
			return newState;
		});
	}

	function init() {
		if (!isInitialized) {
			set(loadFromLocalStorage());
			isInitialized = true;
		}
	}

	return {
		subscribe,
		toggleDaily,
		toggleMonthly,
		toggleLatestTrend,
		init,
		// テスト用のリセット機能
		_reset: () => {
			isInitialized = false;
			set(DEFAULT_STATE);
		},
	};
}

export const summaryVisibility = createSummaryVisibilityStore();
