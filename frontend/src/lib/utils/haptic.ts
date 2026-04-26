/**
 * 触覚フィードバック（バイブレーション）ユーティリティ。
 * web-haptics ライブラリを使用して iOS / Android 両対応の haptic を提供する。
 *
 * iOS (Safari 17.4+):
 *   web-haptics が input[type=checkbox][switch] + label を管理し label.click() で
 *   Taptic Engine を起動する。trigger() の同期部分はユーザージェスチャーコンテキスト
 *   内で実行されるため、pointerdown ハンドラから呼び出すことで iOS でも動作する。
 *
 * Android 等:
 *   Vibration API (navigator.vibrate) を使用する。
 */

import { WebHaptics } from "web-haptics";

// シングルトンインスタンス（DOM 要素はインスタンス内で管理される）
let haptics: WebHaptics | null = null;

function getHaptics(): WebHaptics {
  if (!haptics) {
    haptics = new WebHaptics();
  }
  return haptics;
}

/**
 * ボタン要素に pointerdown ハンドラを付与し、ユーザージェスチャーの中で振動させる。
 * iOS Safari ではユーザージェスチャーコンテキストが必要なため、
 * 非同期の成功コールバックではなくボタン押下時点で発火する。
 * trigger() は async だが hapticLabel.click() は最初の await 前に同期実行されるため
 * ジェスチャーコンテキストが維持される。
 */
export function attachHapticToButton(button: HTMLElement): () => void {
  function handler() {
    // div ラッパーに attach した場合も含め、内包ボタンが disabled なら無視する
    const btn =
      button.tagName === "BUTTON"
        ? (button as HTMLButtonElement)
        : button.querySelector("button");
    if (btn?.disabled) return;
    // void で Promise を明示的に破棄（同期部分のみジェスチャーコンテキストで実行される）
    void getHaptics().trigger("medium");
  }
  button.addEventListener("pointerdown", handler);
  // クリーンアップ関数を返す
  return () => button.removeEventListener("pointerdown", handler);
}
