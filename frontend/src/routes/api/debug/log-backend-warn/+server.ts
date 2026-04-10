import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async () => {
  const response = await fetch("http://backend:8082/debug/warn", {
    method: "GET",
  });
  if (!response.ok) {
    console.error("Error: failed to trigger backend debug warn endpoint");
  }
  return json({ ok: true });
};
