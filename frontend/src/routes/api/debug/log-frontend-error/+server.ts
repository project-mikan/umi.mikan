import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async () => {
  console.error("Error: debug test error triggered from frontend");
  return json({ ok: true });
};
