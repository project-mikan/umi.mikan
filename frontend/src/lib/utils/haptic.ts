/**
 * 触覚フィードバック（バイブレーション）ユーティリティ。
 *
 * iOS Safari (17.4+):
 *   input[type=checkbox][switch] + label.click() を使用して Taptic Engine を起動する。
 *
 *   重要な制約:
 *   - iOS では async 関数内から label.click() を呼び出すと、await より前でも
 *     ユーザージェスチャーコンテキストが失われる場合がある。
 *   - そのため label.click() は pointerdown ハンドラから直接・同期的に呼び出す。
 *   - display:none の要素では Taptic Engine が動作しないため、
 *     position:fixed で画面外に配置して描画ツリーに残す。
 *
 * Android 等:
 *   Vibration API (navigator.vibrate) を使用する。
 */

// iOS向け label 要素（シングルトン）
let iosLabel: HTMLLabelElement | null = null;
let domInitialized = false;

/**
 * iOS向け DOM 要素を初期化する。
 * display:none を使わず position:fixed で画面外に配置して描画ツリーに残す。
 * ユーザージェスチャーの前に呼び出してよい（DOM 生成のみ、haptic は発火しない）。
 */
function initDOM(): void {
  if (domInitialized || typeof document === "undefined") return;
  domInitialized = true;

  // Android / Chrome 系は Vibration API を使用するため DOM 要素不要
  if (
    typeof navigator !== "undefined" &&
    typeof navigator.vibrate === "function"
  ) {
    return;
  }

  const label = document.createElement("label");
  const input = document.createElement("input");
  input.type = "checkbox";
  input.setAttribute("switch", "");
  input.id = "haptic-switch";
  label.setAttribute("for", "haptic-switch");

  // 画面外に配置（display:none は iOS で haptic が動作しないため使用しない）
  label.style.cssText =
    "position:fixed;left:-9999px;top:-9999px;width:1px;height:1px;overflow:hidden;pointer-events:none;";

  label.appendChild(input);
  document.body.appendChild(label);
  iosLabel = label;
}

/**
 * ユーザージェスチャーコンテキスト内で同期的に haptic を発火する。
 * 必ず同期的なイベントハンドラ（pointerdown 等）から直接呼び出すこと。
 * async 関数内から呼び出すと iOS で gesture context が失われる。
 */
function triggerHaptic(): void {
  if (
    typeof navigator !== "undefined" &&
    typeof navigator.vibrate === "function"
  ) {
    // Android / Chrome: Vibration API
    navigator.vibrate(25);
    return;
  }
  // iOS Safari: label.click() でTaptic Engineを起動
  iosLabel?.click();
}

/**
 * ボタン要素に pointerdown ハンドラを付与し、haptic を発火する。
 * iOS Safari ではユーザージェスチャーコンテキストが必要なため、
 * pointerdown で同期的に発火する。
 */
export function attachHapticToButton(button: HTMLElement): () => void {
  // iOS向け DOM 要素を事前生成（ユーザージェスチャーの前に実行してOK）
  initDOM();

  function handler() {
    // div ラッパーに attach した場合も含め、内包ボタンが disabled なら無視する
    const btn =
      button.tagName === "BUTTON"
        ? (button as HTMLButtonElement)
        : button.querySelector("button");
    if (btn?.disabled) return;
    // 同期的に haptic を発火（async 経由にしないことで iOS の gesture context を保持）
    triggerHaptic();
  }

  button.addEventListener("pointerdown", handler);
  // クリーンアップ関数を返す
  return () => button.removeEventListener("pointerdown", handler);
}
