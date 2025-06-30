<script lang="ts">
export let href: string;
export let variant: "primary" | "secondary" = "primary";

const variantClasses = {
	primary: "text-blue-600 hover:text-blue-800",
	secondary: "text-gray-600 hover:text-gray-800",
};

$: classes = `${variantClasses[variant]} font-medium`;
</script>

<a {href} class={classes} {...$$restProps}>
	<slot />
</a>