import { env } from "@/lib/env";

const apiBaseUrl = env.NEXT_PUBLIC_API_URL.replace(/\/+$/, "");

type ApiErrorBody = {
  error?: unknown;
};

export class ApiError extends Error {
  readonly status: number;
  readonly code?: string;

  constructor(status: number, code?: string) {
    super(code ?? `API request failed with status ${status}`);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export function getApiUrl(path: string) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${apiBaseUrl}${normalizedPath}`;
}

export async function apiRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");

  const response = await fetch(getApiUrl(path), {
    ...init,
    headers,
    credentials: "include",
  });

  if (!response.ok) {
    let code: string | undefined;

    try {
      const body = (await response.json()) as ApiErrorBody;
      code = typeof body.error === "string" ? body.error : undefined;
    } catch {
      // The status still provides a useful error when the body is not JSON.
    }

    throw new ApiError(response.status, code);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}
