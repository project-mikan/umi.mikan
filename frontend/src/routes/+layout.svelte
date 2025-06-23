<script lang="ts">
import "../app.css";
import "$lib/i18n";
import type { LayoutData } from "./$types.ts";

export let data: LayoutData;

let isAuthenticated: boolean;
let isAuthPage: boolean;

$: isAuthenticated = data.isAuthenticated;
$: isAuthPage =
	$page.url.pathname === "/login" || $page.url.pathname === "/register";
</script>

<div class="min-h-screen bg-gray-50">
	{#if isAuthenticated && !isAuthPage}
		<nav class="bg-white shadow">
			<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
				<div class="flex h-16 justify-between">
					<div class="flex">
						<div class="flex flex-shrink-0 items-center">
							<h1 class="text-xl font-bold text-gray-900">umi.mikan</h1>
						</div>
					</div>
					<div class="flex items-center">
						<form method="POST" action="?/logout" use:enhance>
							<button
								type="submit"
								class="text-gray-500 hover:text-gray-700 px-3 py-2 text-sm font-medium"
							>
								ログアウト
							</button>
						</form>
					</div>
				</div>
			</div>
		</nav>
	{/if}

	<main class="{isAuthenticated && !isAuthPage ? 'container mx-auto py-8' : ''}">
		<slot />
	</main>
</div>