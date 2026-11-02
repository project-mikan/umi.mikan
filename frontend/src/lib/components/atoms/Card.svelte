<script lang="ts">
  import type { Snippet } from "svelte";

  let { children }: { children?: Snippet } = $props();
</script>

<div
	class="bg-white dark:bg-gray-800 shadow dark:shadow-gray-900/20 rounded-lg text-gray-900 dark:text-gray-100"
>
	{@render children?.()}
</div>
