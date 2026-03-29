/**
 * Entity関連の型定義
 */

/**
 * Entityのカテゴリ
 */
export const EntityCategory = {
	NO_CATEGORY: 0,
	PEOPLE: 1,
} as const;

export type EntityCategoryType =
	(typeof EntityCategory)[keyof typeof EntityCategory];

/**
 * EntityCategoryの表示名を取得
 */
export function getEntityCategoryName(category: EntityCategoryType): string {
	switch (category) {
		case EntityCategory.NO_CATEGORY:
			return "未分類";
		case EntityCategory.PEOPLE:
			return "人物";
		default:
			return "不明";
	}
}

/**
 * Entity位置情報
 */
export interface EntityPosition {
	start: number;
	end: number;
}

/**
 * Entityエイリアス
 */
export interface EntityAlias {
	id: string;
	entityId: string;
	alias: string;
	createdAt: number;
	updatedAt: number;
}

/**
 * Entity
 */
export interface Entity {
	id: string;
	name: string;
	category: EntityCategoryType;
	memo: string;
	aliases: EntityAlias[];
	createdAt: number;
	updatedAt: number;
}
