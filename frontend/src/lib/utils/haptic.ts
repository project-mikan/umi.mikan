/**
 * 触覚フィードバック（バイブレーション）ユーティリティ。
 *
 * iOS (Safari 17.4+):
 *   input[type=checkbox][switch] + label を DOM に常駐させ label.click() で Taptic Engine を起動。
 *   - display:none の要素は iOS でクリックが無効になるため、画面外に配置する方式を使う。
 *   - 非同期コールバック内ではユーザージェスチャーコンテキストが失われる場合があるため、
 *     フォーム送信ボタンに click ハンドラを仕込む方式に変更（下記 attachHapticToButton 参照）。
 *
 * Android 等:
 *   Vibration API (navigator.vibrate) を使用する。
 *   duration: vibrate() に渡すミリ秒数（iOS では無視される）。
 *   20ms は軽いフィードバックとして適切な値。
 */

// iOSデバイス判定
function detectIOS(): boolean {
  if (typeof navigator === "undefined") return false;
  return [/iPhone/i, /iPad/i, /iPod/i].some((re) =>
    re.test(navigator.userAgent),
  );
}

// iOS 向け DOM 要素（画面外に配置して常駐させる）
let hapticLabel: HTMLLabelElement | null = null;

function initHapticElements(): void {
  if (typeof document === "undefined" || hapticLabel) return;

  const input = document.createElement("input");
  input.type = "checkbox";
  input.id = "haptic-switch";
  input.setAttribute("switch", "");
  // display:none では iOS でクリックが無効になるため、画面外に配置する
  input.style.cssText =
    "position:fixed;left:-9999px;top:-9999px;width:1px;height:1px;opacity:0;pointer-events:none;";
  document.body.appendChild(input);

  const label = document.createElement("label");
  label.htmlFor = "haptic-switch";
  label.style.cssText =
    "position:fixed;left:-9999px;top:-9999px;width:1px;height:1px;opacity:0;pointer-events:none;";
  document.body.appendChild(label);
  hapticLabel = label;
}

/**
 * 非同期コールバック内から振動を起こす（Android 向け）。
 * iOS はユーザージェスチャーコンテキストが必要なため attachHapticToButton を使うこと。
 */
export function triggerHaptic(duration = 20): void {
  if (typeof navigator === "undefined") return;

  if (detectIOS()) {
    // iOS はユーザージェスチャー外では label.click() が無効になる場合があるが、
    // フォールバックとして試みる（iOS 17.4+ のみ有効）
    if (!hapticLabel) initHapticElements();
    hapticLabel?.click();
  } else if (navigator.vibrate) {
    navigator.vibrate(duration);
  }
}

/**
 * ボタン要素に pointerdown ハンドラを付与し、ユーザージェスチャーの中で振動させる。
 * iOS Safari ではユーザージェスチャーコンテキストが必要なため、
 * 非同期の成功コールバックではなくボタン押下時点で発火する。
 */
export function attachHapticToButton(
  button: HTMLElement,
  duration = 20,
): () => void {
  function handler() {
    triggerHaptic(duration);
  }
  button.addEventListener("pointerdown", handler);
  // クリーンアップ関数を返す
  return () => button.removeEventListener("pointerdown", handler);
}

// ブラウザ環境では DOMContentLoaded 後に初期化
if (typeof document !== "undefined") {
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initHapticElements, {
      once: true,
    });
  } else {
    initHapticElements();
  }
}
