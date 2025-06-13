import type { PageServerLoad } from './$types';
import { createDiaryClient, promisifyGrpcCall } from '$lib/server/grpc-client';
import * as grpc from '@grpc/grpc-js';
import { error } from '@sveltejs/kit';

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