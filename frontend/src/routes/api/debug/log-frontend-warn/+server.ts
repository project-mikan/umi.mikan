import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async () => {
  console.warn("Warning: debug test warning triggered from frontend");
  return json({ ok: true });
};
