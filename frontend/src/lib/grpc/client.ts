import { createClient } from "@connectrpc/connect";
import { createGrpcWebTransport } from "@connectrpc/connect-web";
import { AuthService } from "./auth/auth_pb.js";
import { DiaryService } from "./diary/diary_pb.js";

let accessToken: string | null = null;

export function setAccessToken(token: string | null) {
	accessToken = token;
}

export function getAccessToken(): string | null {
	return accessToken;
}

const transport = createGrpcWebTransport({
	baseUrl: "http://localhost:8080",
});

const authenticatedTransport = createGrpcWebTransport({
	baseUrl: "http://localhost:8080",
	interceptors: [
		(next) => (req) => {
			if (accessToken) {
				req.header.set("authorization", `Bearer ${accessToken}`);
			}
			return next(req);
		},
	],
});

export const authClient = createClient(AuthService, transport);
export const diaryClient = createClient(DiaryService, authenticatedTransport);
