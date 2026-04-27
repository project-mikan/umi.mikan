/**
 * 触覚フィードバック（バイブレーション）ユーティリティ。
 *
 * iOS Safari (17.4+):
 *   input[type=checkbox][switch] を document.body に position:fixed で配置し、
 *   ボタンの画面座標に合わせてオーバーレイとして表示する。
 *   親要素の pointer-events:none や z-index の影響を受けないようにするため、
 *   document.body 直下に配置する。
 *   ユーザーが直接 UISwitch を触ることで Taptic Engine が起動する。
 *   switch の change イベントで実際のボタンのクリックを発火する。
 *
 * Android 等:
 *   Vibration API (navigator.vibrate) を使用する。
 */

/**
 * クライアントサイドで iOS かどうか判定する。
 * navigator.vibrate が存在しない環境を iOS とみなす。
 */
function detectIOS(): boolean {
  return (
    typeof navigator !== "undefined" && typeof navigator.vibrate !== "function"
  );
}

/**
 * ボタン要素に haptic を付与する。
 *
 * Android:
 *   touchstart / pointerdown で navigator.vibrate(25) を呼び出す。
 *
 * iOS:
 *   ボタンの画面座標に合わせた position:fixed の input[switch] を
 *   document.body 直下に配置する。親要素の CSS の影響を受けないため、
 *   pointer-events:none の親を持つ場合でも正しく動作する。
 *   ボタンの位置変化（キーボード表示等）は ResizeObserver と
 *   visualViewport イベントで追従する。
 *   change イベントで実際の <button> を click() して元のアクションを実行する。
 */
export function attachHapticToButton(button: HTMLElement): () => void {
  if (!detectIOS()) {
    // Android / Chrome: Vibration API
    let lastTime = 0;
    function vibrateHandler() {
      const now = Date.now();
      if (now - lastTime < 50) return;
      lastTime = now;
      const btn =
        button.tagName === "BUTTON"
          ? (button as HTMLButtonElement)
          : button.querySelector("button");
      if (btn?.disabled) return;
      if (
        typeof navigator !== "undefined" &&
        typeof navigator.vibrate === "function"
      ) {
        navigator.vibrate(25);
      }
    }
    button.addEventListener("touchstart", vibrateHandler, { passive: true });
    button.addEventListener("pointerdown", vibrateHandler);
    return () => {
      button.removeEventListener("touchstart", vibrateHandler);
      button.removeEventListener("pointerdown", vibrateHandler);
    };
  }

  // iOS: document.body 直下に position:fixed オーバーレイを配置するアプローチ
  // 実際にクリックすべき <button> 要素を特定する
  const foundBtn =
    button.tagName === "BUTTON"
      ? (button as HTMLButtonElement)
      : button.querySelector<HTMLButtonElement>("button");
  if (!foundBtn) return () => {};
  const actualBtn: HTMLButtonElement = foundBtn;

  // オーバーレイ input を作成する
  const input = document.createElement("input");
  input.type = "checkbox";
  input.setAttribute("switch", "");
  input.setAttribute("tabindex", "-1");
  input.setAttribute("aria-hidden", "true");

  // ボタンの現在の画面座標に合わせてオーバーレイを配置する
  function positionOverlay(): void {
    const rect = actualBtn.getBoundingClientRect();
    // ボタンが非表示（display:none 等）の場合はサイズが 0 になるため
    // そのまま 0x0 の要素になり、タッチできない状態になる
    input.style.cssText = [
      "position:fixed",
      `left:${rect.left}px`,
      `top:${rect.top}px`,
      `width:${rect.width}px`,
      `height:${rect.height}px`,
      "opacity:0",
      "cursor:pointer",
      "appearance:auto",
      "-webkit-appearance:auto",
      "z-index:2147483647",
      "margin:0",
      "padding:0",
      "border:none",
    ].join(";");
  }

  positionOverlay();

  // disabled 状態を同期する（disabled のボタンでは haptic を発火させない）
  input.disabled = actualBtn.disabled;
  const attrObserver = new MutationObserver(() => {
    input.disabled = actualBtn.disabled;
  });
  attrObserver.observe(actualBtn, {
    attributes: true,
    attributeFilter: ["disabled"],
  });

  // ボタンのサイズ変化を監視してオーバーレイ位置を更新する
  const resizeObserver = new ResizeObserver(positionOverlay);
  resizeObserver.observe(actualBtn);

  // iOS キーボード表示時の画面座標変化に追従する
  const vv = window.visualViewport;
  if (vv) {
    vv.addEventListener("resize", positionOverlay, { passive: true });
    vv.addEventListener("scroll", positionOverlay, { passive: true });
  }

  // change イベント: switch の toggle で実際のボタンをクリックする
  function onChange(): void {
    if (!actualBtn.disabled) {
      actualBtn.click();
    }
  }
  input.addEventListener("change", onChange);

  // document.body 直下に配置することで親要素の CSS の影響を受けない
  document.body.appendChild(input);

  return () => {
    attrObserver.disconnect();
    resizeObserver.disconnect();
    if (vv) {
      vv.removeEventListener("resize", positionOverlay);
      vv.removeEventListener("scroll", positionOverlay);
    }
    input.removeEventListener("change", onChange);
    input.remove();
  };
}
