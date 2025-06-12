import type * as grpc from "@grpc/grpc-js";
import type { MessageTypeDefinition } from "@grpc/proto-loader";

import type {
	DiaryServiceClient as _diary_DiaryServiceClient,
	DiaryServiceDefinition as _diary_DiaryServiceDefinition,
} from "./diary/DiaryService";

type SubtypeConstructor<
	Constructor extends new (
		...args: any
	) => any,
	Subtype,
> = {
	new (...args: ConstructorParameters<Constructor>): Subtype;
};

export interface ProtoGrpcType {
	diary: {
		CreateDiaryEntryRequest: MessageTypeDefinition;
		CreateDiaryEntryResponse: MessageTypeDefinition;
		DeleteDiaryEntryRequest: MessageTypeDefinition;
		DeleteDiaryEntryResponse: MessageTypeDefinition;
		DiaryEntry: MessageTypeDefinition;
		DiaryService: SubtypeConstructor<
			typeof grpc.Client,
			_diary_DiaryServiceClient
		> & { service: _diary_DiaryServiceDefinition };
		GetDiaryEntriesByMonthRequest: MessageTypeDefinition;
		GetDiaryEntriesByMonthResponse: MessageTypeDefinition;
		GetDiaryEntriesRequest: MessageTypeDefinition;
		GetDiaryEntriesResponse: MessageTypeDefinition;
		GetDiaryEntryRequest: MessageTypeDefinition;
		GetDiaryEntryResponse: MessageTypeDefinition;
		SearchDiaryEntriesRequest: MessageTypeDefinition;
		SearchDiaryEntriesResponse: MessageTypeDefinition;
		UpdateDiaryEntryRequest: MessageTypeDefinition;
		UpdateDiaryEntryResponse: MessageTypeDefinition;
		YM: MessageTypeDefinition;
		YMD: MessageTypeDefinition;
	};
}
