<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";
import { autoPhraseEnabled } from "$lib/auto-phrase-store";
</script>

<svelte:head>
	<title>{$_("offline.title")} - umi.mikan</title>
</svelte:head>

<div class="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
	<div class="container mx-auto px-4 py-8">
		<div class="max-w-md mx-auto text-center">
			<!-- Offline Icon -->
			<div class="mb-8">
				<svg
					class="w-24 h-24 mx-auto text-gray-400 dark:text-gray-500"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
					xmlns="http://www.w3.org/2000/svg"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M18.364 5.636l-12.728 12.728m0 0L5.636 18.364m12.728-12.728L18.364 18.364M12 12h.01"
					/>
				</svg>
			</div>

			<!-- Title -->
			<h1 class="text-3xl font-bold mb-4 text-gray-900 dark:text-gray-100">
				{$_("offline.title")}
			</h1>

			<!-- Description -->
			<p class="text-lg text-gray-600 dark:text-gray-400 mb-8 {$autoPhraseEnabled ? 'auto-phrase' : ''}">
				{$_("offline.description")}
			</p>

			<!-- Available Features -->
			<div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-6 mb-8 text-left">
				<h2 class="text-xl font-semibold mb-4 text-gray-900 dark:text-gray-100">
					{$_("offline.availableFeatures.title")}
				</h2>
				<ul class="space-y-2 text-gray-700 dark:text-gray-300 {$autoPhraseEnabled ? 'auto-phrase' : ''}">
					<li class="flex items-center">
						<svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
								clip-rule="evenodd"
							/>
						</svg>
						{$_("offline.availableFeatures.readCachedDiaries")}
					</li>
					<li class="flex items-center">
						<svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
								clip-rule="evenodd"
							/>
						</svg>
						{$_("offline.availableFeatures.navigation")}
					</li>
					<li class="flex items-center">
						<svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
								clip-rule="evenodd"
							/>
						</svg>
						{$_("offline.availableFeatures.darkMode")}
					</li>
					<li class="flex items-center">
						<svg class="w-5 h-5 text-green-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
								clip-rule="evenodd"
							/>
						</svg>
						{$_("offline.availableFeatures.languageSettings")}
					</li>
				</ul>
			</div>

			<!-- Limited Features -->
			<div class="bg-red-50 dark:bg-red-900/20 rounded-lg p-6 mb-8 text-left">
				<h2 class="text-xl font-semibold mb-4 text-red-800 dark:text-red-300">
					{$_("offline.limitedFeatures.title")}
				</h2>
				<ul class="space-y-2 text-red-700 dark:text-red-300 {$autoPhraseEnabled ? 'auto-phrase' : ''}">
					<li class="flex items-center">
						<svg class="w-5 h-5 text-red-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
								clip-rule="evenodd"
							/>
						</svg>
						{$_("offline.limitedFeatures.createDiary")}
					</li>
					<li class="flex items-center">
						<svg class="w-5 h-5 text-red-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
								clip-rule="evenodd"
							/>
						</svg>
						{$_("offline.limitedFeatures.editDiary")}
					</li>
					<li class="flex items-center">
						<svg class="w-5 h-5 text-red-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
								clip-rule="evenodd"
							/>
						</svg>
						{$_("offline.limitedFeatures.aiSummary")}
					</li>
					<li class="flex items-center">
						<svg class="w-5 h-5 text-red-500 mr-2" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
								clip-rule="evenodd"
							/>
						</svg>
						{$_("offline.limitedFeatures.search")}
					</li>
				</ul>
			</div>

			<!-- Connection Status -->
			<div class="mb-8">
				<p class="text-sm text-gray-600 dark:text-gray-400 mb-4 {$autoPhraseEnabled ? 'auto-phrase' : ''}">
					{$_("offline.connectionStatus")}
				</p>
				<button
					class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-lg transition-colors"
					on:click={() => window.location.reload()}
				>
					{$_("offline.retryConnection")}
				</button>
			</div>

			<!-- Back to Home -->
			<a
				href="/"
				class="inline-flex items-center text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors"
			>
				<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M10 19l-7-7m0 0l7-7m-7 7h18"
					/>
				</svg>
				{$_("offline.backToHome")}
			</a>
		</div>
	</div>
</div>