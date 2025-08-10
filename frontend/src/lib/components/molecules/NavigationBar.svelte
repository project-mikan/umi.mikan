<script lang="ts">
import { _ } from "svelte-i18n";
import { enhance } from "$app/forms";
import Button from "../atoms/Button.svelte";
import LanguageSelector from "../atoms/LanguageSelector.svelte";
import DarkModeToggle from "../atoms/DarkModeToggle.svelte";

export let isAuthenticated = false;
export let isAuthPage = false;
</script>

{#if isAuthenticated && !isAuthPage}
	<nav class="bg-white dark:bg-gray-900 shadow">
		<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 justify-between">
				<div class="flex">
					<div class="flex flex-shrink-0 items-center">
						<h1 class="text-xl font-bold text-gray-900 dark:text-white">
							<a href="/" class="hover:text-blue-600 dark:hover:text-blue-400"> umi.mikan </a>
						</h1>
					</div>
				</div>
				<div class="flex items-center space-x-4">
					<DarkModeToggle />
					<LanguageSelector />
					<form method="POST" action="?/logout" use:enhance>
						<Button
							type="submit"
							variant="secondary"
							size="sm"
							class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 px-3 py-2"
						>
							{$_("auth.logout")}
						</Button>
					</form>
				</div>
			</div>
		</div>
	</nav>
{:else if isAuthPage}
	<nav class="bg-white dark:bg-gray-900 shadow">
		<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 justify-between">
				<div class="flex">
					<div class="flex flex-shrink-0 items-center">
						<h1 class="text-xl font-bold text-gray-900 dark:text-white">
							<a href="/" class="hover:text-blue-600 dark:hover:text-blue-400"> umi.mikan </a>
						</h1>
					</div>
				</div>
				<div class="flex items-center space-x-4">
					<DarkModeToggle />
					<LanguageSelector />
				</div>
			</div>
		</div>
	</nav>
{/if}

