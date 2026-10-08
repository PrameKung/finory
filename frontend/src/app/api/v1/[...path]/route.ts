import type { NextRequest } from "next/server";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

const hopByHopHeaders = new Set([
  "connection",
  "content-length",
  "host",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
]);

async function proxyToAPI(request: NextRequest) {
  const apiOrigin = process.env.API_PROXY_TARGET;
  if (!apiOrigin) {
    return Response.json({ error: "api_proxy_unavailable" }, { status: 503 });
  }

  let target: URL;
  try {
    target = new URL(`${request.nextUrl.pathname}${request.nextUrl.search}`, apiOrigin);
  } catch {
    return Response.json({ error: "api_proxy_unavailable" }, { status: 503 });
  }

  const headers = new Headers(request.headers);
  for (const name of hopByHopHeaders) headers.delete(name);

  const hasBody = !["GET", "HEAD"].includes(request.method);
  let upstream: Response;
  try {
    upstream = await fetch(target, {
      method: request.method,
      headers,
      body: hasBody ? await request.arrayBuffer() : undefined,
      cache: "no-store",
      redirect: "manual",
    });
  } catch {
    return Response.json({ error: "api_unavailable" }, { status: 502 });
  }

  const responseHeaders = new Headers(upstream.headers);
  responseHeaders.delete("content-encoding");
  responseHeaders.delete("content-length");
  responseHeaders.delete("set-cookie");
  for (const cookie of upstream.headers.getSetCookie()) {
    responseHeaders.append("set-cookie", cookie);
  }

  const body = [204, 205, 304].includes(upstream.status) || request.method === "HEAD"
    ? null
    : await upstream.arrayBuffer();

  return new Response(body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers: responseHeaders,
  });
}

export const GET = proxyToAPI;
export const HEAD = proxyToAPI;
export const POST = proxyToAPI;
export const PUT = proxyToAPI;
export const PATCH = proxyToAPI;
export const DELETE = proxyToAPI;
export const OPTIONS = proxyToAPI;
