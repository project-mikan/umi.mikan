export interface DateInfo {
	year: number;
	month: number;
	day: number;
}

export function getPreviousDate(date: DateInfo): DateInfo {
	const currentDate = new Date(date.year, date.month - 1, date.day);
	currentDate.setDate(currentDate.getDate() - 1);

	return {
		year: currentDate.getFullYear(),
		month: currentDate.getMonth() + 1,
		day: currentDate.getDate(),
	};
}

export function getNextDate(date: DateInfo): DateInfo {
	const currentDate = new Date(date.year, date.month - 1, date.day);
	currentDate.setDate(currentDate.getDate() + 1);

	return {
		year: currentDate.getFullYear(),
		month: currentDate.getMonth() + 1,
		day: currentDate.getDate(),
	};
}

export function formatDateToId(date: DateInfo): string {
	return `${date.year}-${String(date.month).padStart(2, "0")}-${String(date.day).padStart(2, "0")}`;
}

export function getDayOfWeekKey(date: DateInfo): string {
	const jsDate = new Date(date.year, date.month - 1, date.day);
	const dayKeys = [
		"sunday",
		"monday",
		"tuesday",
		"wednesday",
		"thursday",
		"friday",
		"saturday",
	];
	return dayKeys[jsDate.getDay()];
}

export function getPastSameDates(date: DateInfo): {
	oneWeekAgo: DateInfo;
	oneMonthAgo: DateInfo;
	oneYearAgo: DateInfo;
	twoYearsAgo: DateInfo;
} {
	const jsDate = new Date(date.year, date.month - 1, date.day);

	// 1週間前
	const oneWeekAgo = new Date(jsDate);
	oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);

	// 1ヶ月前
	const oneMonthAgo = new Date(jsDate);
	oneMonthAgo.setMonth(oneMonthAgo.getMonth() - 1);

	// 1年前
	const oneYearAgo = new Date(jsDate);
	oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);

	// 2年前
	const twoYearsAgo = new Date(jsDate);
	twoYearsAgo.setFullYear(twoYearsAgo.getFullYear() - 2);

	return {
		oneWeekAgo: {
			year: oneWeekAgo.getFullYear(),
			month: oneWeekAgo.getMonth() + 1,
			day: oneWeekAgo.getDate(),
		},
		oneMonthAgo: {
			year: oneMonthAgo.getFullYear(),
			month: oneMonthAgo.getMonth() + 1,
			day: oneMonthAgo.getDate(),
		},
		oneYearAgo: {
			year: oneYearAgo.getFullYear(),
			month: oneYearAgo.getMonth() + 1,
			day: oneYearAgo.getDate(),
		},
		twoYearsAgo: {
			year: twoYearsAgo.getFullYear(),
			month: twoYearsAgo.getMonth() + 1,
			day: twoYearsAgo.getDate(),
		},
	};
}
