<script lang="ts">
import { autoPhraseEnabled } from "$lib/auto-phrase-store";

export let type: "error" | "success" | "warning" | "info" = "info";

const typeClasses = {
	error:
		"bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300",
	success:
		"bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-300",
	warning:
		"bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 text-yellow-700 dark:text-yellow-300",
	info: "bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300",
};

$: classes = `px-4 py-3 rounded ${typeClasses[type]} ${$autoPhraseEnabled ? "auto-phrase" : ""}`;
</script>

<div class={classes}>
	<slot />
</div>