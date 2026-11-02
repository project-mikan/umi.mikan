<script lang="ts">
  import type { HTMLInputAttributes } from "svelte/elements";

  interface Props extends Omit<HTMLInputAttributes, "type" | "value"> {
    type?: "text" | "email" | "password" | "date";
    value?: string;
    placeholder?: string;
    required?: boolean;
    disabled?: boolean;
    id?: string;
    name?: string;
  }

  let {
    type = "text",
    value = $bindable(""),
    placeholder = "",
    required = false,
    disabled = false,
    id = "",
    name = "",
    autocomplete = "",
    ...rest
  }: Props = $props();

  const baseClasses =
    "block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:focus:ring-blue-400 dark:focus:border-blue-400";
  let classes = $derived(
    `${baseClasses} ${disabled ? "bg-gray-100 dark:bg-gray-800 cursor-not-allowed opacity-50" : ""}`,
  );
</script>

<input
	{type}
	{id}
	{name}
	{placeholder}
	{required}
	{disabled}
	{autocomplete}
	bind:value
	class={classes}
	{...rest}
/>