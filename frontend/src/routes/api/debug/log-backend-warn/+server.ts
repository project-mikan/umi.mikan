import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async () => {
  const response = await fetch("http://backend:8082/debug/warn", {
    method: "GET",
  });
  if (!response.ok) {
    console.error("Error: failed to trigger backend debug warn endpoint");
  }
  return Response.json({ ok: true });
};
