import type { PageServerLoad, Actions } from './$types';
import { createDiaryClient, promisifyGrpcCall } from '$lib/server/grpc-client';
import * as grpc from '@grpc/grpc-js';
import { error, redirect } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ params, cookies }) => {
	const accessToken = cookies.get('accessToken');
	
	if (!accessToken) {
		throw error(401, 'Unauthorized');
	}

	try {
		const client = createDiaryClient();
		const metadata = new grpc.Metadata();
		metadata.add('authorization', `Bearer ${accessToken}`);
		
		const response = await promisifyGrpcCall(
			client,
			'getDiaryEntry',
			{ id: params.id },
			metadata
		);

		if (!response.entry) {
			throw error(404, 'Diary entry not found');
		}

		return {
			entry: response.entry
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
		const dateStr = data.get('date') as string;
		
		if (!content || !dateStr) {
			return {
				error: 'コンテンツと日付は必須です'
			};
		}

		try {
			const date = new Date(dateStr);
			const client = createDiaryClient();
			const metadata = new grpc.Metadata();
			metadata.add('authorization', `Bearer ${accessToken}`);
			
			await promisifyGrpcCall(
				client,
				'updateDiaryEntry',
				{
					id: params.id,
					content,
					date: {
						year: date.getFullYear(),
						month: date.getMonth() + 1,
						day: date.getDate()
					}
				},
				metadata
			);

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
			const client = createDiaryClient();
			const metadata = new grpc.Metadata();
			metadata.add('authorization', `Bearer ${accessToken}`);
			
			await promisifyGrpcCall(
				client,
				'deleteDiaryEntry',
				{ id: params.id },
				metadata
			);

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