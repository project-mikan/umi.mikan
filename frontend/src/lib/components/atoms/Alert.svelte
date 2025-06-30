<script lang="ts">
export let type: "error" | "success" | "warning" | "info" = "info";

const typeClasses = {
	error: "bg-red-50 border border-red-200 text-red-700",
	success: "bg-green-50 border border-green-200 text-green-700",
	warning: "bg-yellow-50 border border-yellow-200 text-yellow-700",
	info: "bg-blue-50 border border-blue-200 text-blue-700",
};

$: classes = `px-4 py-3 rounded ${typeClasses[type]}`;
</script>

<div class={classes}>
	<slot />
</div>