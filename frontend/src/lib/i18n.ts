import { browser } from "$app/environment";
import { init, locale, register, waitLocale } from "svelte-i18n";

register("en", () => import("../locales/en.json"));
register("ja", () => import("../locales/ja.json"));

const initialLocale = browser
	? navigator.language.startsWith("ja")
		? "ja"
		: "en"
	: "en";

init({
	fallbackLocale: "en",
	initialLocale,
	warnOnMissingMessages: false,
});

// Set locale immediately for both SSR and client
locale.set(initialLocale);

export { locale, waitLocale };
