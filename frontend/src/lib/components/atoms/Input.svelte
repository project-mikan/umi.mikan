<script lang="ts">
export let type: "text" | "email" | "password" | "date" = "text";
export let value = "";
export let placeholder = "";
export let required = false;
export let disabled = false;
export let id = "";
export let name = "";
export let autocomplete = "";

const baseClasses =
	"block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500";
$: classes = `${baseClasses} ${disabled ? "bg-gray-100 cursor-not-allowed" : ""}`;
</script>

<input
	{type}
	{id}
	{name}
	{placeholder}
	{required}
	{disabled}
	autocomplete={autocomplete as any}
	bind:value
	class={classes}
	{...$$restProps}
/>