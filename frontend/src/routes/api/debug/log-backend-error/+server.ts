import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async () => {
  const response = await fetch("http://backend:8082/debug/error", {
    method: "GET",
  });
  if (!response.ok) {
    console.error("Error: failed to trigger backend debug error endpoint");
  }
  return json({ ok: true });
};
