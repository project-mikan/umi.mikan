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
