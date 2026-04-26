/**
 * 触覚フィードバック（バイブレーション）ユーティリティ。
 * web-haptics ライブラリを使用して iOS / Android 両対応の haptic を提供する。
 *
 * iOS (Safari 17.4+):
 *   web-haptics が内部で input[type=checkbox][switch] + label を DOM に作成し、
 *   label.click() で Taptic Engine を起動する。
 *
 *   重要な制約:
 *   - web-haptics はデフォルトで要素を display:none にするが、iOS では
 *     display:none の要素から発火した click() では Taptic Engine が動作しない。
 *   - そのため attachHapticToButton の初回呼び出し時に DOM 要素を事前生成し、
 *     display:none → 画面外配置 (position:fixed left:-9999px) に変更する。
 *   - trigger() の同期部分はユーザージェスチャーコンテキスト内で実行されるため、
 *     pointerdown ハンドラから void trigger() で呼び出すことで iOS でも動作する。
 *
 * Android 等:
 *   Vibration API (navigator.vibrate) を使用する。DOM 要素は作成されない。
 */

import { WebHaptics } from "web-haptics";

// シングルトンインスタンス（DOM 要素はインスタンス内で管理される）
let haptics: WebHaptics | null = null;
// DOM 事前初期化済みフラグ（複数コンポーネントから呼ばれても一度だけ実行）
let preInitDone = false;

function getHaptics(): WebHaptics {
  if (!haptics) {
    haptics = new WebHaptics();
  }
  return haptics;
}

/**
 * web-haptics が作成した DOM 要素の display:none を画面外配置に変更する。
 * iOS では display:none の要素への click() では Taptic Engine が動作しないため。
 */
function fixHapticsElementStyles(): void {
  if (typeof document === "undefined") return;
  const label = document.querySelector('label[for^="web-haptics"]');
  if (!label) return; // Android は Vibration API を使用するため DOM 要素なし
  // display:none → 画面外に配置（描画ツリーに残したまま非表示にする）
  (label as HTMLElement).style.cssText =
    "position:fixed;left:-9999px;top:-9999px;width:1px;height:1px;overflow:hidden;";
  const input = label.querySelector("input");
  if (input) {
    // all:initial でリセット後に switch 要素として描画させる
    (input as HTMLInputElement).style.cssText =
      "all:initial;appearance:auto;display:block;";
  }
}

/**
 * DOM 要素を事前初期化する。
 * trigger() を非ジェスチャーコンテキストで呼び出すことで iOS 向け DOM 要素を生成し、
 * その後 display:none を画面外配置に変更する。
 * （非ジェスチャーなので haptic は発火しないが DOM だけ作られる）
 */
function preInitHaptics(): void {
  if (preInitDone || typeof document === "undefined") return;
  preInitDone = true;
  // trigger() の同期部分が実行されることで iOS 向け DOM 要素が生成される
  void getHaptics().trigger("light");
  // trigger() の同期部分完了後（await 前）に DOM 要素が存在するのでスタイルを修正
  fixHapticsElementStyles();
}

/**
 * ボタン要素に pointerdown ハンドラを付与し、ユーザージェスチャーの中で振動させる。
 * iOS Safari ではユーザージェスチャーコンテキストが必要なため、
 * 非同期の成功コールバックではなくボタン押下時点で発火する。
 */
export function attachHapticToButton(button: HTMLElement): () => void {
  // 初回呼び出し時に iOS 向け DOM 要素を事前初期化してスタイルを修正する
  preInitHaptics();

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
