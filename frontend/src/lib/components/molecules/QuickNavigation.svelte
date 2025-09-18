<script lang="ts">
import { _ } from "svelte-i18n";
import "$lib/i18n";

const currentDate = new Date();
const currentYear = currentDate.getFullYear();
const currentMonth = currentDate.getMonth() + 1;
</script>

<nav class="quick-nav fixed z-50 
	md:hidden
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
	</a>
	
	<a 
		href="/monthly/{currentYear}/{currentMonth}"
		class="nav-button flex items-center justify-center p-3 rounded-md bg-green-500 hover:bg-green-600 text-white transition-colors duration-200 min-w-[44px] min-h-[44px]"
		aria-label={$_("navigation.monthly")}
	>
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
		</svg>
	</a>
	
	<a
		href="/search"
		class="nav-button flex items-center justify-center p-3 rounded-md bg-purple-500 hover:bg-purple-600 text-white transition-colors duration-200 min-w-[44px] min-h-[44px]"
		aria-label={$_("navigation.search")}
	>
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
		</svg>
	</a>

	<a
		href="/settings"
		class="nav-button flex items-center justify-center p-3 rounded-md bg-orange-500 hover:bg-orange-600 text-white transition-colors duration-200 min-w-[44px] min-h-[44px]"
		aria-label={$_("navigation.settings")}
	>
		<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
		</svg>
	</a>
</nav>

<style>
.quick-nav {
	max-width: calc(100vw - 2rem);
}

.nav-button {
	flex: 1;
	max-width: 80px;
}
</style>