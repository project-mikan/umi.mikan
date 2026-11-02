<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLAnchorAttributes } from "svelte/elements";

  interface Props extends HTMLAnchorAttributes {
    href: string;
    variant?: "primary" | "secondary";
    children?: Snippet;
  }

  let { href, variant = "primary", children, ...rest }: Props = $props();

  const variantClasses = {
    primary:
      "text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300",
    secondary:
      "text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200",
  };

  let classes = $derived(`${variantClasses[variant]} font-medium`);
</script>

<a {href} class={classes} {...rest}>
	{@render children?.()}
</a>