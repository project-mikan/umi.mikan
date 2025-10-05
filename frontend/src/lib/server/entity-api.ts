import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import {
	CreateEntityAliasRequestSchema,
	type CreateEntityAliasResponse,
	CreateEntityRequestSchema,
	type CreateEntityResponse,
	DeleteEntityAliasRequestSchema,
	type DeleteEntityAliasResponse,
	DeleteEntityRequestSchema,
	type DeleteEntityResponse,
	type EntityCategory,
	EntityService,
	GetDiariesByEntityRequestSchema,
	type GetDiariesByEntityResponse,
	GetEntityRequestSchema,
	type GetEntityResponse,
	ListEntitiesRequestSchema,
	type ListEntitiesResponse,
	SearchEntitiesRequestSchema,
	type SearchEntitiesResponse,
	UpdateEntityRequestSchema,
	type UpdateEntityResponse,
} from "$lib/grpc/entity/entity_pb";

/**
 * 認証付きgRPCトランスポートを作成
 */
function createAuthenticatedTransport(accessToken: string) {
	return createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${accessToken}`);
				return next(req);
			},
		],
	});
}

export interface CreateEntityParams {
	name: string;
	category: EntityCategory;
	memo: string;
	accessToken: string;
}

export interface UpdateEntityParams {
	id: string;
	name: string;
	category: EntityCategory;
	memo: string;
	accessToken: string;
}

export interface DeleteEntityParams {
	id: string;
	accessToken: string;
}

export interface GetEntityParams {
	id: string;
	accessToken: string;
}

export interface ListEntitiesParams {
	category: EntityCategory;
	allCategories: boolean;
	accessToken: string;
}

export interface CreateEntityAliasParams {
	entityId: string;
	alias: string;
	accessToken: string;
}

export interface DeleteEntityAliasParams {
	id: string;
	accessToken: string;
}

export interface GetDiariesByEntityParams {
	entityId: string;
	accessToken: string;
}

export interface SearchEntitiesParams {
	query: string;
	accessToken: string;
}

/**
 * エンティティを作成
 */
export async function createEntity(
	params: CreateEntityParams,
): Promise<CreateEntityResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(EntityService, transport);

	const request = create(CreateEntityRequestSchema, {
		name: params.name,
		category: params.category,
		memo: params.memo,
	});

	return await client.createEntity(request);
}

/**
 * エンティティを更新
 */
export async function updateEntity(
	params: UpdateEntityParams,
): Promise<UpdateEntityResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(EntityService, transport);

	const request = create(UpdateEntityRequestSchema, {
		id: params.id,
		name: params.name,
		category: params.category,
		memo: params.memo,
	});

	return await client.updateEntity(request);
}

/**
 * エンティティを削除
 */
export async function deleteEntity(
	params: DeleteEntityParams,
): Promise<DeleteEntityResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(EntityService, transport);

	const request = create(DeleteEntityRequestSchema, {
		id: params.id,
	});

	return await client.deleteEntity(request);
}

/**
 * エンティティを取得
 */
export async function getEntity(
	params: GetEntityParams,
): Promise<GetEntityResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(EntityService, transport);

	const request = create(GetEntityRequestSchema, {
		id: params.id,
	});

	return await client.getEntity(request);
}

/**
 * エンティティ一覧を取得
 */
export async function listEntities(
	params: ListEntitiesParams,
): Promise<ListEntitiesResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(EntityService, transport);

	const request = create(ListEntitiesRequestSchema, {
		category: params.category,
		allCategories: params.allCategories,
	});

	return await client.listEntities(request);
}

/**
 * エイリアスを追加
 */
export async function createEntityAlias(
	params: CreateEntityAliasParams,
): Promise<CreateEntityAliasResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(EntityService, transport);

	const request = create(CreateEntityAliasRequestSchema, {
		entityId: params.entityId,
		alias: params.alias,
	});

	return await client.createEntityAlias(request);
}

/**
 * エイリアスを削除
 */
export async function deleteEntityAlias(
	params: DeleteEntityAliasParams,
): Promise<DeleteEntityAliasResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(EntityService, transport);

	const request = create(DeleteEntityAliasRequestSchema, {
		id: params.id,
	});

	return await client.deleteEntityAlias(request);
}

/**
 * エンティティに紐づく日記を取得
 */
export async function getDiariesByEntity(
	params: GetDiariesByEntityParams,
): Promise<GetDiariesByEntityResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(EntityService, transport);

	const request = create(GetDiariesByEntityRequestSchema, {
		entityId: params.entityId,
	});

	return await client.getDiariesByEntity(request);
}

/**
 * エンティティを検索（ユーザーの入力に対する候補表示）
 */
export async function searchEntities(
	params: SearchEntitiesParams,
): Promise<SearchEntitiesResponse> {
	const transport = createAuthenticatedTransport(params.accessToken);
	const client = createClient(EntityService, transport);

	const request = create(SearchEntitiesRequestSchema, {
		query: params.query,
	});

	return await client.searchEntities(request);
}
