<script lang="ts">
	import { _ } from "svelte-i18n";
	import { enhance } from "$app/forms";
	import { dev } from "$app/environment";
	import Button from "../atoms/Button.svelte";
	import LanguageSelector from "../atoms/LanguageSelector.svelte";
	import DarkModeToggle from "../atoms/DarkModeToggle.svelte";
	import BudouXToggle from "../atoms/BudouXToggle.svelte";

	export let isAuthenticated = false;
	export let isAuthPage = false;

	const currentDate = new Date();
	const currentYear = currentDate.getFullYear();
	const currentMonth = currentDate.getMonth() + 1;
</script>

{#if isAuthenticated && !isAuthPage}
	<header class="sticky top-0 z-50 bg-white dark:bg-gray-900 shadow">
		<nav class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 justify-between">
				<div class="flex items-center">
					<div class="flex flex-shrink-0 items-center">
						<h1 class="text-xl font-bold text-gray-900 dark:text-white">
							<a href="/" class="hover:text-blue-600 dark:hover:text-blue-400"> umi.mikan </a>
						</h1>
					</div>
					<!-- Desktop Navigation -->
					<div class="hidden md:flex md:ml-8 md:space-x-4">
						<a
							href="/"
							class="flex items-center px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-blue-600 dark:text-gray-200 dark:hover:text-blue-400 transition-colors"
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path>
							</svg>
							{$_("navigation.home")}
						</a>
						<a
							href="/monthly/{currentYear}/{currentMonth}"
							class="flex items-center px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-green-600 dark:text-gray-200 dark:hover:text-green-400 transition-colors"
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
							</svg>
							{$_("navigation.monthly")}
						</a>
						<a
							href="/search"
							class="flex items-center px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-purple-600 dark:text-gray-200 dark:hover:text-purple-400 transition-colors"
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
							</svg>
							{$_("navigation.search")}
						</a>
						<a
							href="/entities"
							class="flex items-center px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-yellow-600 dark:text-gray-200 dark:hover:text-yellow-400 transition-colors"
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
							</svg>
							{$_("navigation.entities")}
						</a>
						<a
							href="/settings"
							class="flex items-center px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-orange-600 dark:text-gray-200 dark:hover:text-orange-400 transition-colors"
						>
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
							</svg>
							{$_("navigation.settings")}
						</a>
						{#if dev}
							<a
								href="/debug"
								class="flex items-center px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:text-red-600 dark:text-gray-200 dark:hover:text-red-400 transition-colors"
							>
								<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"></path>
								</svg>
								{$_("navigation.debug")}
							</a>
						{/if}
					</div>
				</div>
				<div class="flex items-center space-x-4">
					<BudouXToggle />
					<DarkModeToggle />
					<LanguageSelector />
				</div>
			</div>
		</nav>
	</header>
{:else if isAuthPage}
	<header class="sticky top-0 z-50 bg-white dark:bg-gray-900 shadow">
		<nav class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 justify-between">
				<div class="flex">
					<div class="flex flex-shrink-0 items-center">
						<h1 class="text-xl font-bold text-gray-900 dark:text-white">
							<a href="/" class="hover:text-blue-600 dark:hover:text-blue-400"> umi.mikan </a>
						</h1>
					</div>
				</div>
				<div class="flex items-center space-x-4">
					<BudouXToggle />
					<DarkModeToggle />
					<LanguageSelector />
				</div>
			</div>
		</nav>
	</header>
{/if}

