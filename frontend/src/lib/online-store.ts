import { readable } from "svelte/store";
import { browser } from "$app/environment";

// ネットワーク接続状態を管理するストア
export const isOnline = readable<boolean>(true, (set) => {
  if (!browser) return;

  // 初期値をnavigator.onLineから取得
  set(navigator.onLine);

  const handleOnline = () => set(true);
  const handleOffline = () => set(false);

  window.addEventListener("online", handleOnline);
  window.addEventListener("offline", handleOffline);

  return () => {
    window.removeEventListener("online", handleOnline);
    window.removeEventListener("offline", handleOffline);
  };
});
