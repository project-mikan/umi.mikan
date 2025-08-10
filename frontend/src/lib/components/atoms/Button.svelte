<script lang="ts">
export let type: "button" | "submit" = "button";
export let variant: "primary" | "secondary" | "danger" | "success" | "gray" =
	"primary";
export let size: "sm" | "md" | "lg" = "md";
export let disabled = false;
export let loading = false;

const baseClasses =
	"font-medium rounded-md focus:outline-none focus:ring-2 focus:ring-offset-2 dark:focus:ring-offset-gray-800 transition-colors";

const variantClasses = {
	primary: "bg-blue-600 hover:bg-blue-700 text-white focus:ring-blue-500",
	secondary:
		"border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 focus:ring-indigo-500",
	danger: "bg-red-600 hover:bg-red-700 text-white focus:ring-red-500",
	success: "bg-green-600 hover:bg-green-700 text-white focus:ring-green-500",
	gray: "bg-gray-600 hover:bg-gray-700 dark:bg-gray-500 dark:hover:bg-gray-400 text-white focus:ring-gray-500",
};

const sizeClasses = {
	sm: "px-3 py-1.5 text-sm",
	md: "px-4 py-2 text-sm",
	lg: "px-6 py-3 text-base",
};

$: classes = `${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${disabled || loading ? "opacity-50 cursor-not-allowed" : ""}`;
</script>

<button {type} class={classes} {disabled} on:click {...$$restProps}>
	<slot />
</button>
