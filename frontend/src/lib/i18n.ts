import { init, register, locale, waitLocale } from 'svelte-i18n';

register('en', () => import('../locales/en.json'));
register('ja', () => import('../locales/ja.json'));

init({
	fallbackLocale: 'en',
	initialLocale: typeof window !== 'undefined' ? 
		navigator.language.startsWith('ja') ? 'ja' : 'en' : 'en'
});

export { locale, waitLocale };