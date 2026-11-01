import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async () => {
  console.warn("Warning: debug test warning triggered from frontend");
  return Response.json({ ok: true });
};
