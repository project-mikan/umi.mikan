<script lang="ts">
  interface Props {
    htmlFor?: string;
    required?: boolean;
    srOnly?: boolean;
    children?: import("svelte").Snippet;
  }

  let {
    htmlFor = "",
    required = false,
    srOnly = false,
    children,
  }: Props = $props();

  let classes = $derived(
    `${srOnly ? "sr-only" : "block text-sm font-medium text-gray-700 dark:text-gray-300"}`,
  );
</script>

<label for={htmlFor} class={classes}>
	{@render children?.()}
	{#if required && !srOnly}
		<span class="text-red-500 dark:text-red-400">*</span>
	{/if}
</label>