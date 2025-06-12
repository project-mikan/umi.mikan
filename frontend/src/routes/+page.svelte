<script lang="ts">
import { goto } from "$app/navigation";
import { redirect } from "@sveltejs/kit";
import { onMount } from "svelte";
import type { PageData } from "./$types";

export let data: PageData;

onMount(() => {
	if (!data.isAuthenticated) {
		goto("/login");
	}
});
</script>

{#if data.isAuthenticated}
	<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="py-8">
			<h1 class="text-3xl font-bold text-gray-900 mb-8">日記アプリ</h1>
			<div class="bg-white shadow rounded-lg p-6">
				<p class="text-gray-600">認証が完了しました。日記機能の実装をお待ちください。</p>
			</div>
		</div>
	</div>
{:else}
	<div class="flex items-center justify-center min-h-screen">
		<div class="text-center">
			<p class="text-gray-600">ログインページにリダイレクトしています...</p>
		</div>
	</div>
{/if}
