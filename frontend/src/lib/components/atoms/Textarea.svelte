<script lang="ts">
  import { createEventDispatcher, onDestroy, onMount } from "svelte";
  import type { DiaryHighlight } from "$lib/types/highlight";
  import {
    getTextOffset,
    restoreCursorPosition,
  } from "$lib/utils/cursor-utils";
  import { highlightEntitiesAndHighlights } from "$lib/utils/diary-entity-highlighter";
  import { htmlToPlainText } from "$lib/utils/html-text-converter";

  export let value = "";
  export let placeholder = "";
  export let required = false;
  export let disabled = false;
  export let id = "";
  export let name = "";
  export let rows = 4;
  export let diaryHighlights: DiaryHighlight[] = [];
  export let searchKeyword = "";

  // ハイライトを適用すべき元のコンテンツ
  let savedContent = "";

  const dispatch = createEventDispatcher();

  let contentElement: HTMLDivElement;
  let isUpdatingFromValue = false; // valueからの更新中かどうかのフラグ
  let isComposing = false; // IME入力中かどうかのフラグ
  let updateTimeout: ReturnType<typeof setTimeout> | null = null; // ハイライト更新のタイムアウト
  let isTyping = false; // ユーザーが入力中かどうかのフラグ

  // diaryHighlightsが変更されたときもsavedContentを更新
  // 配列の内容をシリアライズして比較（参照比較では毎回新しい配列なので不十分）
  let previousDiaryHighlightsKey = "";
  $: {
    // valueとdiaryHighlightsへの依存を明示（初期化タイミングを確実にする）
    void value;

    // diaryHighlightsの内容をキーに変換（長さ + 最初の要素の情報）
    const currentKey = diaryHighlights
      ? `${diaryHighlights.length}-${diaryHighlights[0]?.start ?? ""}-${diaryHighlights[0]?.end ?? ""}`
      : "empty";

    if (currentKey !== previousDiaryHighlightsKey) {
      previousDiaryHighlightsKey = currentKey;
      savedContent = value;
    }
  }

  // ハイライトを適用したHTMLを生成
  // 入力中でない場合のみハイライトを表示
  // また、現在のvalueが保存されたコンテンツと一致する場合のみハイライトを適用
  $: highlightedHTML = (() => {
    // 入力中は常にプレーンテキスト
    if (isTyping) {
      return value.replace(/\n/g, "<br>");
    }

    const hasSearchKeyword = searchKeyword && searchKeyword.trim().length > 0;
    const hasDiaryHighlights = diaryHighlights && diaryHighlights.length > 0;

    // ハイライト情報が何もない場合はプレーンテキスト
    if (!hasDiaryHighlights && !hasSearchKeyword) {
      return value.replace(/\n/g, "<br>");
    }

    // diaryHighlightsを適用する場合は保存済みコンテンツと一致が必要
    // （編集中のテキストには古いpositionデータを適用しない）
    const highlightsToApply =
      hasDiaryHighlights && value === savedContent ? diaryHighlights : [];

    // ハイライトを適用
    return highlightEntitiesAndHighlights(
      value,
      [],
      highlightsToApply,
      hasSearchKeyword ? searchKeyword.trim() : undefined,
    );
  })();

  // captureフェーズでTabキーをキャプチャするためのリスナー
  onMount(async () => {
    // savedContentを確実に初期化（SSR時のリアクティビティの問題を回避）
    if (!savedContent) {
      savedContent = value;
    }

    if (contentElement) {
      // captureフェーズで追加してTabキーを早期にキャプチャ
      contentElement.addEventListener("keydown", _handleKeydown, true);
      // previousHighlightedHTMLを同期してリアクティブブロックの二重呼び出しを防ぐ
      previousHighlightedHTML = highlightedHTML;
      updateContentElement();
    }
  });

  onDestroy(() => {
    if (contentElement) {
      contentElement.removeEventListener("keydown", _handleKeydown, true);
    }
  });

  const baseClasses =
    "block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 rounded-md shadow-sm focus:outline-none resize-none min-h-24 whitespace-pre-wrap [&>br]:leading-none [&>br]:h-0";
  $: classes = `${baseClasses} ${disabled ? "bg-gray-100 dark:bg-gray-800 cursor-not-allowed opacity-50" : ""}`;

  // Calculate min height based on rows
  $: minHeight = `${rows * 1.5}rem`;

  // highlightedHTMLが変化したときにDOMを更新する
  // searchKeyword・diaryHighlights・value・isTypingの変化を一元的に処理する
  let previousHighlightedHTML = "";
  $: {
    if (contentElement && !isUpdatingFromValue && !isComposing && !isTyping) {
      if (highlightedHTML !== previousHighlightedHTML) {
        previousHighlightedHTML = highlightedHTML;
        updateContentElement();
      }
    }
  }

  function updateContentElement() {
    if (!contentElement) return;
    // SSR時は何もしない
    if (typeof window === "undefined") return;

    // 現在のDOMの内容が、これから設定しようとする内容と実質的に同じ場合は
    // innerHTMLの再代入をスキップする。
    // ※例えば文末でEnterを押した直後は、_handleKeydownが既にDOMへ直接<br>を
    //   挿入済みで見た目上は改行されている。その後isTypingが500ms後にfalseへ
    //   戻ったタイミングでこの関数が呼ばれると、highlightedHTMLの中身（改行を
    //   含むvalue由来のHTML）は既存のDOMと同じであるにもかかわらず、
    //   innerHTMLを再代入してDOM要素を作り直してしまい、その一瞬の再構築が
    //   「改行が一瞬戻る」ようなちらつきとして見える原因になっていた。
    //   カーソルアンカー用のゼロ幅スペースはDOM側にのみ存在しうるため、
    //   比較前に取り除く。
    const currentHtmlWithoutAnchor = contentElement.innerHTML.replace(/​/g, "");
    if (currentHtmlWithoutAnchor === highlightedHTML) {
      return;
    }

    // 現在のカーソル位置を取得
    const selection = window.getSelection();
    let cursorPos = 0;

    if (selection && selection.rangeCount > 0) {
      const range = selection.getRangeAt(0);
      cursorPos = getTextOffset(
        contentElement,
        range.startContainer,
        range.startOffset,
      );
    }

    // HTMLを更新
    contentElement.innerHTML = highlightedHTML;

    // カーソル位置を復元
    if (cursorPos > 0) {
      restoreCursorPosition(contentElement, cursorPos);
    }
  }

  function _handleInput(event: Event) {
    const target = event.target as HTMLDivElement;
    isUpdatingFromValue = true;
    isTyping = true; // 入力中フラグを立てる

    // 既存のタイムアウトをクリア
    if (updateTimeout !== null) {
      clearTimeout(updateTimeout);
    }

    // IME入力中（compositionupdate）の場合は、valueの更新のみ行う
    const isCompositionUpdate = event instanceof CompositionEvent;

    value = htmlToPlainText(target.innerHTML);

    // contentElementが初期化されていない場合は何もしない
    if (!contentElement) {
      isUpdatingFromValue = false;
      return;
    }

    if (isCompositionUpdate) {
      isUpdatingFromValue = false;

      // 500ms後に入力が止まったら入力完了フラグを下ろす
      updateTimeout = setTimeout(() => {
        isTyping = false;
      }, 500);

      return;
    }

    isUpdatingFromValue = false;

    // 500ms後に入力が止まったら入力完了フラグを下ろす
    // （ハイライトは保存後にサーバーから返されるpositionデータで適用される）
    updateTimeout = setTimeout(() => {
      isTyping = false; // 入力完了フラグ
    }, 500);
  }

  function _handleKeydown(event: KeyboardEvent) {
    if (event.ctrlKey && event.key === "Enter") {
      event.preventDefault();
      dispatch("save");
    } else if (event.key === "Enter") {
      // Ignore Enter key during IME composition (Japanese input)
      if (event.isComposing) {
        return;
      }

      // Prevent default behavior and manually insert <br>
      // This handles both Enter and Shift+Enter
      event.preventDefault();

      // Insert a <br> tag at the cursor position
      const selection = window.getSelection();
      if (selection && selection.rangeCount > 0) {
        const range = selection.getRangeAt(0);
        const br = document.createElement("br");

        // Delete any selected content first
        range.deleteContents();

        // カーソルの直前のノードをチェック
        // 直前のノードがBRタグの場合、連続改行とみなす
        let previousNode = range.startContainer;
        if (
          previousNode.nodeType === Node.TEXT_NODE &&
          range.startOffset === 0
        ) {
          // テキストノードの先頭にいる場合、前の兄弟ノードをチェック
          previousNode = previousNode.previousSibling as Node;
        } else if (previousNode.nodeType === Node.ELEMENT_NODE) {
          // 要素ノード内にいる場合、startOffsetで指定された位置の直前の子ノードをチェック
          const children = previousNode.childNodes;
          if (range.startOffset > 0 && range.startOffset <= children.length) {
            previousNode = children[range.startOffset - 1];
          }
        }

        const isPreviousNodeBR = previousNode && previousNode.nodeName === "BR";

        // Insert the br element
        range.insertNode(br);

        // contentEditableの末尾での改行を処理するために、
        // <br>の後ろに意味のあるコンテンツが存在するかチェック
        // 末尾の場合のみ、もう一つ<br>を挿入してカーソルが次の行に表示されるようにする
        // ただし、直前のノードがBRタグの場合（連続改行）は、1つのbrのみを挿入
        function hasContentAfterNode(node: Node): boolean {
          let current: Node | null = node.nextSibling;
          while (current) {
            // テキストノードの場合
            if (current.nodeType === Node.TEXT_NODE) {
              const text = current.textContent || "";
              // 空白以外のテキストがある場合はコンテンツありとみなす
              if (text.trim().length > 0) {
                return true;
              }
              // 空白のみの場合は次のノードをチェック
            }
            // 要素ノードの場合
            else if (current.nodeType === Node.ELEMENT_NODE) {
              const element = current as Element;
              // BRタグ以外の要素があればコンテンツありとみなす
              if (element.nodeName !== "BR") {
                return true;
              }
              // BRタグの場合は次のノードをチェック（連続するBRタグを全て無視）
            }
            current = current.nextSibling;
          }
          // 親ノードがcontentElementでない場合、親の兄弟をチェック
          if (node.parentNode && node.parentNode !== contentElement) {
            return hasContentAfterNode(node.parentNode);
          }
          return false;
        }

        const isAtEnd = !hasContentAfterNode(br);

        // カーソルを<br>の直後に置く際、コンテナ要素基準のoffset（例: setStartAfter(br)）で
        // Rangeを設定すると、ブラウザによっては直後のキー入力が<br>の"前"ではなく"後ろ"に
        // 挿入されてしまうことがある（Chromeで確認済み）。これを避けるため、<br>の直後に
        // ゼロ幅スペース（U+200B）を1文字持つテキストノードを挿入し、そのテキストノード内
        // （offset 1、＝ゼロ幅スペースの直後）にカーソルを置く。空文字列のテキストノードだと
        // ブラウザが正規化時に削除してしまいカーソル位置が失われることがあったため、
        // 削除されない1文字のゼロ幅スペースを使う。ゼロ幅スペースはhtmlToPlainText側で
        // 除去されるためvalueには反映されない。
        const cursorAnchor = document.createTextNode("​");
        br.after(cursorAnchor);

        // 直前のノードがBRタグでない かつ 末尾の場合のみ、カーソル表示用に一時的な2つ目の<br>を挿入
        // ※この2つ目の<br>はDOM表示上カーソルを次行に見せるためだけのものであり、
        //   value（プレーンテキスト）には反映してはいけない。
        //   htmlToPlainTextは<br>1つにつき\nを1つ生成するため、2つ挿入したまま
        //   input発火するとEnter1回でvalueに空行が2行分（\n\n）入ってしまう。
        //   cursorAnchorへのカーソル位置はafterBrの追加・削除の影響を受けない
        //   （テキストノード自体は増減しないため）。
        let afterBr: HTMLBRElement | null = null;
        if (!isPreviousNodeBR && isAtEnd) {
          afterBr = document.createElement("br");
          cursorAnchor.after(afterBr);
        }

        // カーソルをcursorAnchor（ゼロ幅スペースを持つテキストノード）内、
        // ゼロ幅スペースの直後（offset 1）に配置
        const newRange = document.createRange();
        newRange.setStart(cursorAnchor, 1);
        newRange.collapse(true);
        selection.removeAllRanges();
        selection.addRange(newRange);

        // valueへの反映用に、表示専用の一時的な2つ目の<br>は変換前に取り除く
        if (afterBr) {
          afterBr.remove();
        }
      }

      // Trigger input event to update the value
      const inputEvent = new Event("input", { bubbles: true });
      contentElement.dispatchEvent(inputEvent);
      // isTypingはdispatchによって呼ばれた_handleInput内のsetTimeoutに任せる
      // ここで即座にisTyping = falseにするとhighlightedHTMLの変化でupdateContentElement()が
      // 呼ばれてカーソル位置が上書きされてしまう
    }
  }

  // フォーカスが外れた時の自動保存
  function _handleBlur() {
    // IME入力中は自動保存しない
    if (isComposing) return;
    dispatch("autosave");
  }
</script>

<!-- Hidden input for form submission -->
<input type="hidden" {name} {value} {required} />

<div class="relative">
	<div
		bind:this={contentElement}
		{id}
		data-placeholder={placeholder}
		contenteditable={!disabled}
		class="{classes} auto-phrase-target"
		style="min-height: {minHeight}; line-height: 18pt; font-size:11pt; font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre;white-space:pre-wrap;padding: 4px;"
		on:input={_handleInput}
		on:blur={_handleBlur}
		on:compositionstart={() => { isComposing = true; }}
		on:compositionupdate={_handleInput}
		on:compositionend={(event) => {
			isComposing = false;
			// IME確定後に処理を実行
			_handleInput(event);
		}}
		{...$$restProps}
	></div>
</div>
