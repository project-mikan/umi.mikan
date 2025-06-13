import type { Actions } from './$types';
import { createDiaryClient, promisifyGrpcCall } from '$lib/server/grpc-client';
import * as grpc from '@grpc/grpc-js';
import { error, redirect } from '@sveltejs/kit';

export const actions: Actions = {
	default: async ({ request, cookies }) => {
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
				'createDiaryEntry',
				{
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
			console.error('Failed to create diary entry:', err);
			return {
				error: '日記の作成に失敗しました'
			};
		}
	}
};