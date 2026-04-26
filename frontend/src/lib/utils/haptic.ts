/**
 * 触覚フィードバック（バイブレーション）ユーティリティ。
 *
 * iOS Safari (17.4+):
 *   ユーザーのタッチが物理的に input[type=checkbox][switch] に当たるよう、
 *   ボタン上に透明なオーバーレイとして配置する。
 *   プログラマティックな click() ではなく、ユーザーが直接 UISwitch を触ることで
 *   Taptic Engine が起動する（iOS 26 でのプログラマティック haptic 制限に対応）。
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
 *   ボタン上に透明な input[switch] オーバーレイを配置する。
 *   ボタンが <div> ラッパーの場合: div 内に inset:0 で配置する。
 *   ボタンが <button> 要素の場合: 親要素に position:relative を設定し、
 *   button の offsetLeft/Top/Width/Height に合わせた位置に配置する。
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

  // iOS: オーバーレイアプローチ
  // 実際にクリックすべき <button> 要素を特定する
  const foundBtn =
    button.tagName === "BUTTON"
      ? (button as HTMLButtonElement)
      : button.querySelector<HTMLButtonElement>("button");
  if (!foundBtn) return () => {};
  const actualBtn: HTMLButtonElement = foundBtn;

  // オーバーレイを配置するコンテナを決定する
  // <div> ラッパーの場合: ラッパー自身をコンテナにして inset:0 で覆う
  // <button> 要素の場合: 親要素をコンテナにして button の位置・サイズに合わせる
  const isDirectButton = button.tagName === "BUTTON";
  const container = isDirectButton
    ? (button.parentElement as HTMLElement)
    : button;
  if (!container) return () => {};

  // コンテナに position:relative がなければ設定する（absolute 配置の基準にする）
  const savedPosition = container.style.position;
  if (getComputedStyle(container).position === "static") {
    container.style.position = "relative";
  }

  // オーバーレイ input を作成する
  const input = document.createElement("input");
  input.type = "checkbox";
  input.setAttribute("switch", "");
  input.setAttribute("tabindex", "-1");
  input.setAttribute("aria-hidden", "true");
  input.style.cssText =
    "position:absolute;opacity:0;cursor:pointer;appearance:auto;-webkit-appearance:auto;z-index:1;margin:0;";

  // 位置・サイズを設定する
  if (isDirectButton) {
    // <button> の場合: 親内での offsetLeft/Top/Width/Height に合わせる
    input.style.left = button.offsetLeft + "px";
    input.style.top = button.offsetTop + "px";
    input.style.width = button.offsetWidth + "px";
    input.style.height = button.offsetHeight + "px";
  } else {
    // <div> ラッパーの場合: ラッパー全体を覆う
    input.style.left = "0";
    input.style.top = "0";
    input.style.right = "0";
    input.style.bottom = "0";
  }

  // disabled 状態を同期する（disabled のボタンでは haptic を発火させない）
  input.disabled = actualBtn.disabled;
  const attrObserver = new MutationObserver(() => {
    input.disabled = actualBtn.disabled;
  });
  attrObserver.observe(actualBtn, {
    attributes: true,
    attributeFilter: ["disabled"],
  });

  // change イベント: switch の toggle で実際のボタンをクリックする
  function onChange(): void {
    if (!actualBtn.disabled) {
      actualBtn.click();
    }
  }
  input.addEventListener("change", onChange);
  container.appendChild(input);

  return () => {
    attrObserver.disconnect();
    input.removeEventListener("change", onChange);
    input.remove();
    // コンテナの position スタイルを元に戻す
    if (isDirectButton) {
      container.style.position = savedPosition;
    }
  };
}
