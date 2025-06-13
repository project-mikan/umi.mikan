import type { PageServerLoad } from './$types';
import { getDiaryEntry } from '$lib/server/diary-api';
import { error } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const accessToken = cookies.get('accessToken');
	
	if (!accessToken) {
		throw error(401, 'Unauthorized');
	}

	try {
		// params.id should be in format YYYY-MM-DD
		const dateMatch = params.id.match(/^(\d{4})-(\d{2})-(\d{2})$/);
		if (!dateMatch) {
			throw error(400, 'Invalid date format');
		}

		const [, year, month, day] = dateMatch;
		const entry = await getDiaryEntry({
			year: parseInt(year, 10),
			month: parseInt(month, 10),
			day: parseInt(day, 10)
		});

		if (!entry) {
			throw error(404, 'Diary entry not found');
		}

		return {
			entry
		};
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}
		console.error('Failed to load diary entry:', err);
		throw error(500, 'Failed to load diary entry');
	}
};