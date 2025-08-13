// Utility function to format diary count messages
export function formatDiaryCountMessage(
	count: number,
	locale: "ja" | "en" = "en",
): string {
	const messages = {
		ja: "これまでに{count}日分の日記を書きました",
		en: "You have written {count} diary entries so far",
	};

	return messages[locale].replace("{count}", count.toString());
}

// Date formatting utility used in the page component
export function formatDateStr(ymd: {
	year: number;
	month: number;
	day: number;
}): string {
	return `${ymd.year}-${String(ymd.month).padStart(2, "0")}-${String(ymd.day).padStart(2, "0")}`;
}

// Monthly URL generation utility
export function getMonthlyUrl(date: Date = new Date()): string {
	return `/monthly/${date.getFullYear()}/${date.getMonth() + 1}`;
}
