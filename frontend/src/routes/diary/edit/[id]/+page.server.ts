import type { PageServerLoad, Actions } from './$types';
import { getDiaryEntry, updateDiaryEntry, deleteDiaryEntry } from '$lib/server/diary-api';
import { error, redirect } from '@sveltejs/kit';

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

export const actions: Actions = {
	update: async ({ request, params, cookies }) => {
		const accessToken = cookies.get('accessToken');
		
		if (!accessToken) {
			throw error(401, 'Unauthorized');
		}

		const data = await request.formData();
		const content = data.get('content') as string;
		const title = data.get('title') as string || '';
		const dateStr = data.get('date') as string;
		
		if (!content || !dateStr) {
			return {
				error: 'コンテンツと日付は必須です'
			};
		}

		try {
			// First, get the current entry to get the ID
			const dateMatch = params.id.match(/^(\d{4})-(\d{2})-(\d{2})$/);
			if (!dateMatch) {
				throw error(400, 'Invalid date format');
			}

			const [, year, month, day] = dateMatch;
			const currentEntry = await getDiaryEntry({
				year: parseInt(year, 10),
				month: parseInt(month, 10),
				day: parseInt(day, 10)
			});

			if (!currentEntry) {
				throw error(404, 'Diary entry not found');
			}

			const date = new Date(dateStr);
			await updateDiaryEntry({
				id: currentEntry.id,
				title,
				content,
				date: {
					year: date.getFullYear(),
					month: date.getMonth() + 1,
					day: date.getDate()
				}
			});

			throw redirect(303, '/diary');
		} catch (err) {
			if (err instanceof Response) {
				throw err;
			}
			console.error('Failed to update diary entry:', err);
			return {
				error: '日記の更新に失敗しました'
			};
		}
	},

	delete: async ({ params, cookies }) => {
		const accessToken = cookies.get('accessToken');
		
		if (!accessToken) {
			throw error(401, 'Unauthorized');
		}

		try {
			// First, get the current entry to get the ID
			const dateMatch = params.id.match(/^(\d{4})-(\d{2})-(\d{2})$/);
			if (!dateMatch) {
				throw error(400, 'Invalid date format');
			}

			const [, year, month, day] = dateMatch;
			const currentEntry = await getDiaryEntry({
				year: parseInt(year, 10),
				month: parseInt(month, 10),
				day: parseInt(day, 10)
			});

			if (!currentEntry) {
				throw error(404, 'Diary entry not found');
			}

			await deleteDiaryEntry(currentEntry.id);

			throw redirect(303, '/diary');
		} catch (err) {
			if (err instanceof Response) {
				throw err;
			}
			console.error('Failed to delete diary entry:', err);
			return {
				error: '日記の削除に失敗しました'
			};
		}
	}
};