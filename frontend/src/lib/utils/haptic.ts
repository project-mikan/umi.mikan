/**
 * use-haptic ライブラリ（https://github.com/posaune0423/use-haptic）と同じ方式で実装。
 *
 * iOS では input[type=checkbox][switch] + label を DOM に常駐させ、
 * label.click() で Taptic Engine を起動する。
 * 要素をクリックごとに生成・破棄すると iOS で動作しないため、
 * モジュールロード時に一度だけ作成して使い回す（use-haptic の useEffect 相当）。
 *
 * Android 等では Vibration API を使用する。
 * duration: Android の vibrate() に渡すミリ秒数（iOS では無視される）。
 * 20ms は軽いフィードバックとして適切な値。
 */

// iOSデバイス判定（use-haptic/esm/utils.js の detectiOS と同一ロジック）
function detectIOS(): boolean {
  if (typeof navigator === "undefined") return false;
  return [/iPhone/i, /iPad/i, /iPod/i].some((re) =>
    re.test(navigator.userAgent),
  );
}

// iOS 向け DOM 要素をモジュールロード時に一度だけ初期化（use-haptic の useEffect 相当）
let hapticInput: HTMLInputElement | null = null;
let hapticLabel: HTMLLabelElement | null = null;

function initHapticElements(): void {
  if (typeof document === "undefined" || hapticInput) return;

  const input = document.createElement("input");
  input.type = "checkbox";
  input.id = "haptic-switch";
  input.setAttribute("switch", "");
  input.style.display = "none";
  document.body.appendChild(input);
  hapticInput = input;

  const label = document.createElement("label");
  label.htmlFor = "haptic-switch";
  label.style.display = "none";
  document.body.appendChild(label);
  hapticLabel = label;
}

export function triggerHaptic(duration = 20): void {
  if (typeof document === "undefined") return;

  if (detectIOS()) {
    // DOM 要素が未初期化なら作成（初回呼び出し時のフォールバック）
    initHapticElements();
    hapticLabel?.click();
  } else if (navigator?.vibrate) {
    navigator.vibrate(duration);
  }
}

// ブラウザ環境では +layout.svelte の onMount より先にモジュールが評価されるため、
// document.body が存在するタイミングを保証するため遅延初期化とする
if (typeof document !== "undefined") {
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initHapticElements, {
      once: true,
    });
  } else {
    initHapticElements();
  }
}
