<script lang="ts">
import type { PageData } from "./$types.ts";

export let data: PageData;
</script>

{#if data.isAuthenticated}
	<div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="py-8">
			<h1 class="text-3xl font-bold text-gray-900 mb-8">日記アプリ</h1>
			<div class="bg-white shadow rounded-lg p-6">
				<p class="text-gray-600 mb-6">日記アプリへようこそ！</p>
				<div class="space-y-4">
					<a 
						href="/diary" 
						class="inline-block bg-blue-600 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded"
					>
						日記一覧を見る
					</a>
					<a 
						href="/diary/create" 
						class="inline-block ml-4 bg-green-600 hover:bg-green-700 text-white font-bold py-2 px-4 rounded"
					>
						新しい日記を書く
					</a>
				</div>
			</div>
		</div>
	</div>
{:else}
	<div class="flex items-center justify-center min-h-screen">
		<div class="text-center">
			<h1 class="text-4xl font-bold text-gray-900 mb-8">日記アプリ</h1>
			<p class="text-gray-600 mb-8">あなたの日々の思い出を記録しましょう</p>
			<div class="space-x-4">
				<a 
					href="/login" 
					class="inline-block bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-6 rounded-lg transition duration-200"
				>
					ログイン
				</a>
				<a 
					href="/register" 
					class="inline-block bg-green-600 hover:bg-green-700 text-white font-bold py-3 px-6 rounded-lg transition duration-200"
				>
					新規登録
				</a>
			</div>
		</div>
	</div>
{/if}
