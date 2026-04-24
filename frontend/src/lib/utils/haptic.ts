/**
 * iOSデバイス判定
 */
function detectIOS(): boolean {
  if (typeof navigator === "undefined") return false;
  return [/iPhone/i, /iPad/i, /iPod/i].some((re) =>
    re.test(navigator.userAgent),
  );
}

/**
 * 保存時などに触覚フィードバック（バイブレーション）を発生させる。
 * - iOS: input[switch] 要素を利用
 * - Android 等: Vibration API を利用
 */
export function triggerHaptic(duration = 5): void {
  if (typeof document === "undefined") return;

  if (detectIOS()) {
    // iOSはinput[type=checkbox][switch]のlabelクリックで触覚フィードバック
    const input = document.createElement("input");
    input.type = "checkbox";
    input.setAttribute("switch", "");
    input.style.display = "none";
    const id = `haptic-${Date.now()}`;
    input.id = id;
    document.body.appendChild(input);

    const label = document.createElement("label");
    label.htmlFor = id;
    label.style.display = "none";
    document.body.appendChild(label);

    label.click();

    // クリーンアップ
    setTimeout(() => {
      document.body.removeChild(input);
      document.body.removeChild(label);
    }, 100);
  } else if (navigator?.vibrate) {
    navigator.vibrate(duration);
  }
}
