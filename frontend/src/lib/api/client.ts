import { env } from "@/lib/env";
import { ApiError, parseApiErrorBody } from "@/lib/api/errors";

export { ApiError, getApiErrorMessage } from "@/lib/api/errors";

const apiBaseUrl = env.NEXT_PUBLIC_API_URL.replace(/\/+$/, "");

async function readApiError(response: Response) {
  try {
    return parseApiErrorBody(await response.json());
  } catch {
    return {};
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

  let response: Response;

  try {
    response = await fetch(getApiUrl(path), {
      ...init,
      headers,
      credentials: "include",
    });
  } catch (error) {
    if (init.signal?.aborted) {
      throw error;
    }

    throw new ApiError(0, "network_error", { cause: error });
  }

  if (!response.ok) {
    const { code, message } = await readApiError(response);

    throw new ApiError(response.status, code, { message });
  }

  if (response.status === 204) {
    return undefined as T;
  }

  try {
    return (await response.json()) as T;
  } catch (error) {
    throw new ApiError(response.status, "invalid_response", { cause: error });
  }
}
