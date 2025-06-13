import type { PageServerLoad } from './$types';
import { createDiaryClient, promisifyGrpcCall } from '$lib/server/grpc-client';
import * as grpc from '@grpc/grpc-js';
import { error } from '@sveltejs/kit';

export const load: PageServerLoad = async ({ cookies }) => {
	const accessToken = cookies.get('accessToken');
	
	if (!accessToken) {
		throw error(401, 'Unauthorized');
	}

	try {
		const client = createDiaryClient();
		const metadata = new grpc.Metadata();
		metadata.add('authorization', `Bearer ${accessToken}`);
		
		// 今月の日記エントリを取得
		const now = new Date();
		const response = await promisifyGrpcCall(
			client,
			'getDiaryEntriesByMonth',
			{
				month: {
					year: now.getFullYear(),
					month: now.getMonth() + 1
				}
			},
			metadata
		);

		return {
			entries: response.entries || []
		};
	} catch (err) {
		console.error('Failed to load diary entries:', err);
		return {
			entries: []
		};
	}
};