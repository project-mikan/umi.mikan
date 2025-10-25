<script lang="ts">
	export let href: string;
	export let variant: "primary" | "secondary" = "primary";

	const variantClasses = {
		primary:
			"text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300",
		secondary:
			"text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200",
	};

	$: classes = `${variantClasses[variant]} font-medium`;
</script>

<a {href} class={classes} {...$$restProps}>
	<slot />
</a>