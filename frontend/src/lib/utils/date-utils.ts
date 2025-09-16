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
	twoMonthsAgo: DateInfo;
	sixMonthsAgo: DateInfo;
	oneYearAgo: DateInfo;
	twoYearsAgo: DateInfo;
	threeYearsAgo: DateInfo;
	fourYearsAgo: DateInfo;
	fiveYearsAgo: DateInfo;
	sixYearsAgo: DateInfo;
	sevenYearsAgo: DateInfo;
	eightYearsAgo: DateInfo;
	nineYearsAgo: DateInfo;
	tenYearsAgo: DateInfo;
} {
	const jsDate = new Date(date.year, date.month - 1, date.day);

	// 1週間前
	const oneWeekAgo = new Date(jsDate);
	oneWeekAgo.setDate(oneWeekAgo.getDate() - 7);

	// 1ヶ月前
	const oneMonthAgo = new Date(jsDate);
	oneMonthAgo.setMonth(oneMonthAgo.getMonth() - 1);

	// 2ヶ月前
	const twoMonthsAgo = new Date(jsDate);
	twoMonthsAgo.setMonth(twoMonthsAgo.getMonth() - 2);

	// 半年前
	const sixMonthsAgo = new Date(jsDate);
	sixMonthsAgo.setMonth(sixMonthsAgo.getMonth() - 6);

	// 1年前
	const oneYearAgo = new Date(jsDate);
	oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);

	// 2年前
	const twoYearsAgo = new Date(jsDate);
	twoYearsAgo.setFullYear(twoYearsAgo.getFullYear() - 2);

	// 3年前
	const threeYearsAgo = new Date(jsDate);
	threeYearsAgo.setFullYear(threeYearsAgo.getFullYear() - 3);

	// 4年前
	const fourYearsAgo = new Date(jsDate);
	fourYearsAgo.setFullYear(fourYearsAgo.getFullYear() - 4);

	// 5年前
	const fiveYearsAgo = new Date(jsDate);
	fiveYearsAgo.setFullYear(fiveYearsAgo.getFullYear() - 5);

	// 6年前
	const sixYearsAgo = new Date(jsDate);
	sixYearsAgo.setFullYear(sixYearsAgo.getFullYear() - 6);

	// 7年前
	const sevenYearsAgo = new Date(jsDate);
	sevenYearsAgo.setFullYear(sevenYearsAgo.getFullYear() - 7);

	// 8年前
	const eightYearsAgo = new Date(jsDate);
	eightYearsAgo.setFullYear(eightYearsAgo.getFullYear() - 8);

	// 9年前
	const nineYearsAgo = new Date(jsDate);
	nineYearsAgo.setFullYear(nineYearsAgo.getFullYear() - 9);

	// 10年前
	const tenYearsAgo = new Date(jsDate);
	tenYearsAgo.setFullYear(tenYearsAgo.getFullYear() - 10);

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
		twoMonthsAgo: {
			year: twoMonthsAgo.getFullYear(),
			month: twoMonthsAgo.getMonth() + 1,
			day: twoMonthsAgo.getDate(),
		},
		sixMonthsAgo: {
			year: sixMonthsAgo.getFullYear(),
			month: sixMonthsAgo.getMonth() + 1,
			day: sixMonthsAgo.getDate(),
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
		threeYearsAgo: {
			year: threeYearsAgo.getFullYear(),
			month: threeYearsAgo.getMonth() + 1,
			day: threeYearsAgo.getDate(),
		},
		fourYearsAgo: {
			year: fourYearsAgo.getFullYear(),
			month: fourYearsAgo.getMonth() + 1,
			day: fourYearsAgo.getDate(),
		},
		fiveYearsAgo: {
			year: fiveYearsAgo.getFullYear(),
			month: fiveYearsAgo.getMonth() + 1,
			day: fiveYearsAgo.getDate(),
		},
		sixYearsAgo: {
			year: sixYearsAgo.getFullYear(),
			month: sixYearsAgo.getMonth() + 1,
			day: sixYearsAgo.getDate(),
		},
		sevenYearsAgo: {
			year: sevenYearsAgo.getFullYear(),
			month: sevenYearsAgo.getMonth() + 1,
			day: sevenYearsAgo.getDate(),
		},
		eightYearsAgo: {
			year: eightYearsAgo.getFullYear(),
			month: eightYearsAgo.getMonth() + 1,
			day: eightYearsAgo.getDate(),
		},
		nineYearsAgo: {
			year: nineYearsAgo.getFullYear(),
			month: nineYearsAgo.getMonth() + 1,
			day: nineYearsAgo.getDate(),
		},
		tenYearsAgo: {
			year: tenYearsAgo.getFullYear(),
			month: tenYearsAgo.getMonth() + 1,
			day: tenYearsAgo.getDate(),
		},
	};
}
