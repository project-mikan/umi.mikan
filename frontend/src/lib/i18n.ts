import { init, locale, register, waitLocale } from "svelte-i18n";
import { browser } from "$app/environment";

register("en", () => import("../locales/en.json"));
register("ja", () => import("../locales/ja.json"));

init({
	fallbackLocale: "en",
	initialLocale: browser
		? navigator.language.startsWith("ja")
			? "ja"
			: "en"
		: "en",
});

export { locale, waitLocale };
