<script lang="ts">
export let padding: "sm" | "md" | "lg" = "md";

const paddingClasses = {
	sm: "p-4",
	md: "p-6",
	lg: "p-8",
};

$: classes = `bg-white shadow rounded-lg ${paddingClasses[padding]}`;
</script>

<div class={classes}>
	<slot />
</div>