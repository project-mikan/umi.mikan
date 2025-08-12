<script lang="ts">
export let padding: "sm" | "md" | "lg" = "md";

const paddingClasses = {
	sm: "p-4",
	md: "p-6",
	lg: "p-8",
};

$: classes = `bg-white dark:bg-gray-800 shadow dark:shadow-gray-900/20 rounded-lg text-gray-900 dark:text-gray-100 ${paddingClasses[padding]}`;
</script>

<div class={classes}>
	<slot />
</div>

