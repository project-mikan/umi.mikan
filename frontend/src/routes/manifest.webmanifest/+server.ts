import type { RequestHandler } from "./$types";

// Manifest configuration for different languages
const getManifestConfig = (lang: string = "en") => {
	const manifestConfigs = {
		ja: {
			name: "umi.mikan - 日記アプリ",
			short_name: "umi.mikan",
			description: "毎日使う日記アプリ",
			lang: "ja",
		},
		en: {
			name: "umi.mikan - Diary App",
			short_name: "umi.mikan",
			description: "Daily diary application",
			lang: "en",
		},
	};

	const selectedConfig =
		manifestConfigs[lang as keyof typeof manifestConfigs] || manifestConfigs.en;

	return {
		...selectedConfig,
		start_url: "/",
		scope: "/",
		display: "standalone",
		background_color: "#ffffff",
		theme_color: "#3b82f6",
		orientation: "portrait-primary",
		categories: ["lifestyle", "productivity"],
		id: "/",
		prefer_related_applications: false,
		icons: [
			{
				src: "icons/icon-72x72.png",
				sizes: "72x72",
				type: "image/png",
			},
			{
				src: "icons/icon-96x96.png",
				sizes: "96x96",
				type: "image/png",
			},
			{
				src: "icons/icon-128x128.png",
				sizes: "128x128",
				type: "image/png",
			},
			{
				src: "icons/icon-144x144.png",
				sizes: "144x144",
				type: "image/png",
			},
			{
				src: "icons/icon-152x152.png",
				sizes: "152x152",
				type: "image/png",
			},
			{
				src: "icons/icon-192x192.png",
				sizes: "192x192",
				type: "image/png",
			},
			{
				src: "icons/icon-384x384.png",
				sizes: "384x384",
				type: "image/png",
			},
			{
				src: "icons/icon-512x512.png",
				sizes: "512x512",
				type: "image/png",
			},
			{
				src: "icons/icon-192x192.png",
				sizes: "192x192",
				type: "image/png",
				purpose: "maskable",
			},
			{
				src: "icons/icon-512x512.png",
				sizes: "512x512",
				type: "image/png",
				purpose: "maskable",
			},
		],
	};
};

export const GET: RequestHandler = async ({ url, cookies }) => {
	// Get language from query parameter, cookie, or default to English
	const langParam = url.searchParams.get("lang");
	const langCookie = cookies.get("locale");
	const lang = langParam || langCookie || "en";

	const manifest = getManifestConfig(lang);

	return new Response(JSON.stringify(manifest), {
		headers: {
			"Content-Type": "application/manifest+json",
			"Cache-Control": "public, max-age=3600", // Cache for 1 hour
		},
	});
};
