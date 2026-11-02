import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async () => {
  console.error("Error: debug test error triggered from frontend");
  return Response.json({ ok: true });
};
