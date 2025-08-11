<script lang="ts">
import { onMount } from "svelte";
import { _ } from "svelte-i18n";
import "$lib/i18n";

let yearProgress = 0;
let monthProgress = 0;
let dayProgress = 0;

function calculateProgress() {
	const now = new Date();

	// 今年の経過日数
	const startOfYear = new Date(now.getFullYear(), 0, 1);
	const endOfYear = new Date(now.getFullYear(), 11, 31);
	const yearTotal =
		(endOfYear.getTime() - startOfYear.getTime()) / (1000 * 60 * 60 * 24) + 1;
	const yearElapsed =
		(now.getTime() - startOfYear.getTime()) / (1000 * 60 * 60 * 24) + 1;
	yearProgress = (yearElapsed / yearTotal) * 100;

	// 今月の経過日数
	const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
	const endOfMonth = new Date(now.getFullYear(), now.getMonth() + 1, 0);
	const monthTotal = endOfMonth.getDate();
	const monthElapsed = now.getDate();
	monthProgress = (monthElapsed / monthTotal) * 100;

	// 今日の経過時間（時：分）
	const startOfDay = new Date(now.getFullYear(), now.getMonth(), now.getDate());
	const dayElapsed = now.getTime() - startOfDay.getTime();
	const dayTotal = 24 * 60 * 60 * 1000;
	dayProgress = (dayElapsed / dayTotal) * 100;
}

onMount(() => {
	calculateProgress();
	const interval = setInterval(calculateProgress, 60000);
	return () => clearInterval(interval);
});
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg p-4 shadow">
	<div class="space-y-3">
		<div>
			<div class="flex justify-between items-center mb-1">
				<span class="text-sm font-medium text-gray-700 dark:text-gray-300"
					>{$_("timeProgress.dayProgress")}</span
				>
				<span class="text-sm text-gray-500 dark:text-gray-400"
					>{dayProgress.toFixed(1)}%</span
				>
			</div>
			<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
				<div
					class="bg-orange-600 h-2 rounded-full transition-all duration-500"
					style="width: {dayProgress}%"
				></div>
			</div>
		</div>

		<div>
			<div class="flex justify-between items-center mb-1">
				<span class="text-sm font-medium text-gray-700 dark:text-gray-300"
					>{$_("timeProgress.monthProgress")}</span
				>
				<span class="text-sm text-gray-500 dark:text-gray-400"
					>{monthProgress.toFixed(1)}%</span
				>
			</div>
			<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
				<div
					class="bg-green-600 h-2 rounded-full transition-all duration-500"
					style="width: {monthProgress}%"
				></div>
			</div>
		</div>

		<div>
			<div class="flex justify-between items-center mb-1">
				<span class="text-sm font-medium text-gray-700 dark:text-gray-300"
					>{$_("timeProgress.yearProgress")}</span
				>
				<span class="text-sm text-gray-500 dark:text-gray-400"
					>{yearProgress.toFixed(1)}%</span
				>
			</div>
			<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
				<div
					class="bg-blue-600 h-2 rounded-full transition-all duration-500"
					style="width: {yearProgress}%"
				></div>
			</div>
		</div>
	</div>
</div>

