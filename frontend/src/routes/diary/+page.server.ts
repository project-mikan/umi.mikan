import type { PageServerLoad } from './$types';
import { getDiaryEntriesByMonth } from '$lib/server/diary-api';
import { error } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ cookies }) => {
	const accessToken = cookies.get('accessToken');
	
	if (!accessToken) {
		throw error(401, 'Unauthorized');
	}

	try {
		// 今月の日記エントリを取得
		const now = new Date();
		const entries = await getDiaryEntriesByMonth({
			year: now.getFullYear(),
			month: now.getMonth() + 1
		});

		return {
			entries
		};
	} catch (err) {
		console.error('Failed to load diary entries:', err);
		return {
			entries: []
		};
	}
};