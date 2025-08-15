<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";

const currentDate = new Date();
const currentYear = currentDate.getFullYear();
const currentMonth = currentDate.getMonth() + 1;
</script>

<nav class="quick-nav fixed z-50 
	md:static md:z-auto md:flex-row md:w-full md:justify-center md:translate-x-0 md:bg-gray-50 md:dark:bg-gray-800 md:shadow-none md:border-none
	bottom-4 left-1/2 -translate-x-1/2 flex-row w-max
	flex gap-2 p-2 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-600"
>
	<a 
		href="/"
		class="nav-button flex items-center justify-center p-3 rounded-md bg-blue-500 hover:bg-blue-600 text-white transition-colors duration-200 min-w-[44px] min-h-[44px]"
		aria-label={$_("navigation.home")}
	>
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path>
		</svg>
		<span class="hidden md:block ml-2 text-sm">{$_("navigation.home")}</span>
	</a>
	
	<a 
		href="/monthly/{currentYear}/{currentMonth}"
		class="nav-button flex items-center justify-center p-3 rounded-md bg-green-500 hover:bg-green-600 text-white transition-colors duration-200 min-w-[44px] min-h-[44px]"
		aria-label={$_("navigation.monthly")}
	>
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
		</svg>
		<span class="hidden md:block ml-2 text-sm">{$_("navigation.monthly")}</span>
	</a>
	
	<a 
		href="/search"
		class="nav-button flex items-center justify-center p-3 rounded-md bg-purple-500 hover:bg-purple-600 text-white transition-colors duration-200 min-w-[44px] min-h-[44px]"
		aria-label={$_("navigation.search")}
	>
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
		</svg>
		<span class="hidden md:block ml-2 text-sm">{$_("navigation.search")}</span>
	</a>
</nav>

<style>
@media (max-width: 768px) {
	.quick-nav {
		max-width: calc(100vw - 2rem);
	}
	
	.nav-button {
		flex: 1;
		max-width: 80px;
	}
}

@media (min-width: 768px) {
	.nav-button {
		width: 140px;
		justify-content: flex-start;
	}
}
</style>