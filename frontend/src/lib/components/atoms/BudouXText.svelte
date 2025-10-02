<script lang="ts">
import { applyBudouX } from "$lib/utils/budoux";

/** 表示するテキスト */
export let text: string;

// BudouXで処理されたHTMLを生成
$: processedHtml = applyBudouX(text);
</script>

<!-- eslint-disable-next-line svelte/no-at-html-tags -->
{@html processedHtml}
