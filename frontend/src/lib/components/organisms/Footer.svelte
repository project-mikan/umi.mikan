<script lang="ts">
import { _ } from "svelte-i18n";
import { enhance } from "$app/forms";
import Button from "../atoms/Button.svelte";
import Link from "../atoms/Link.svelte";

export let isAuthenticated = false;
</script>

{#if isAuthenticated}
	<footer class="bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 pb-16 sm:pb-6">
		<div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
			<div class="flex flex-col items-center space-y-4">
				<div>
					<Link 
						href="https://github.com/project-mikan/umi.mikan" 
						target="_blank"
						class="text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100"
					>
						{$_("footer.github")}
					</Link>
				</div>
				<div>
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
	</footer>
{/if}