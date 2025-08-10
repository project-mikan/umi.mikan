<script lang="ts">
import { _ } from "svelte-i18n";
import { enhance } from "$app/forms";
import Button from "../atoms/Button.svelte";

export let isAuthenticated = false;
export let isAuthPage = false;
</script>

{#if isAuthenticated && !isAuthPage}
	<nav class="bg-white shadow">
		<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 justify-between">
				<div class="flex">
					<div class="flex flex-shrink-0 items-center">
						<h1 class="text-xl font-bold text-gray-900">
							<a href="/" class="hover:text-blue-600"> umi.mikan </a>
						</h1>
					</div>
				</div>
				<div class="flex items-center">
					<form method="POST" action="?/logout" use:enhance>
						<Button
							type="submit"
							variant="secondary"
							size="sm"
							class="text-gray-500 hover:text-gray-700 px-3 py-2"
						>
							{$_("auth.logout")}
						</Button>
					</form>
				</div>
			</div>
		</div>
	</nav>
{/if}

