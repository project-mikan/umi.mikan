/**
 * 触覚フィードバック（バイブレーション）ユーティリティ。
 *
 * iOS Safari (17.4+):
 *   input[type=checkbox][switch] + label.click() を使用して Taptic Engine を起動する。
 *
 *   重要な制約:
 *   - iOS では async 関数内から label.click() を呼び出すと、await より前でも
 *     ユーザージェスチャーコンテキストが失われる場合がある。
 *   - そのため label.click() は touchstart / pointerdown ハンドラから直接・同期的に呼び出す。
 *   - display:none の要素では Taptic Engine が動作しないため、
 *     opacity:0 で非表示にして描画ツリーに残す。
 *   - input に appearance:auto を明示しないと UISwitch として描画されず Taptic Engine が起動しない。
 *   - touchstart と pointerdown の両方が発火するタッチデバイスでは、
 *     直近の発火から 50ms 以内の重複呼び出しをスキップして二重振動を防ぐ。
 *
 * Android 等:
 *   Vibration API (navigator.vibrate) を使用する。
 */

// iOS向け label 要素（シングルトン）
let iosLabel: HTMLLabelElement | null = null;
let domInitialized = false;

// 直近の haptic 発火時刻（タッチデバイスでの二重発火防止）
let lastHapticTime = 0;

/**
 * iOS向け DOM 要素を初期化する。
 * display:none を使わず opacity:0 で描画ツリーに残す。
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

  // display:none は iOS で haptic が動作しないため使用しない。
  // opacity:0 で非表示にして描画ツリーに残す。
  // pointer-events:none は label.click() に干渉する可能性があるため使用しない。
  label.style.cssText =
    "position:fixed;left:-9999px;top:0;width:1px;height:1px;overflow:hidden;opacity:0;";
  // iOS が UISwitch として描画するために appearance:auto を明示する。
  // これがないと通常のチェックボックスとして扱われ、Taptic Engine が起動しない。
  input.style.cssText = "appearance:auto;-webkit-appearance:auto;";

  label.appendChild(input);
  document.body.appendChild(label);
  iosLabel = label;
}

/**
 * ユーザージェスチャーコンテキスト内で同期的に haptic を発火する。
 * 必ず同期的なイベントハンドラ（touchstart / pointerdown 等）から直接呼び出すこと。
 * async 関数内から呼び出すと iOS で gesture context が失われる。
 */
function triggerHaptic(): void {
  // 直近 50ms 以内に発火済みなら二重振動を防ぐ
  const now = Date.now();
  if (now - lastHapticTime < 50) return;
  lastHapticTime = now;

  if (
    typeof navigator !== "undefined" &&
    typeof navigator.vibrate === "function"
  ) {
    // Android / Chrome: Vibration API
    navigator.vibrate(25);
    return;
  }
  // iOS Safari: label.click() で Taptic Engine を起動
  iosLabel?.click();
}

/**
 * ボタン要素に touchstart / pointerdown ハンドラを付与し、haptic を発火する。
 * touchstart は iOS で最も確実なジェスチャーコンテキストであり先に発火する。
 * pointerdown はマウスデバイス向けのフォールバック。
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

  // touchstart: iOS で最も確実なジェスチャーコンテキスト
  button.addEventListener("touchstart", handler, { passive: true });
  // pointerdown: マウス・ペンデバイス向けフォールバック
  button.addEventListener("pointerdown", handler);
  // クリーンアップ関数を返す
  return () => {
    button.removeEventListener("touchstart", handler);
    button.removeEventListener("pointerdown", handler);
  };
}
