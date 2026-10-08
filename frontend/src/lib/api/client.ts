import { ApiError, parseApiErrorBody } from "./errors.ts";

export { ApiError, getApiErrorMessage } from "./errors.ts";

const refreshPath = "/api/v1/auth/refresh";

let refreshPromise: Promise<void> | undefined;

async function readApiError(response: Response) {
  try {
    return parseApiErrorBody(await response.json());
  } catch {
    return {};
  }
}

export function getApiUrl(path: string) {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return normalizedPath;
}

async function performApiRequest<T>(
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

function refreshAccessToken() {
  refreshPromise ??= performApiRequest<void>(refreshPath, {
    method: "POST",
  }).finally(() => {
    refreshPromise = undefined;
  });

  return refreshPromise;
}

export async function apiRequest<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  try {
    return await performApiRequest<T>(path, init);
  } catch (error) {
    if (
      path === refreshPath ||
      !(error instanceof ApiError) ||
      error.status !== 401
    ) {
      throw error;
    }

    try {
      await refreshAccessToken();
    } catch (refreshError) {
      if (refreshError instanceof ApiError && refreshError.status === 401) {
        throw error;
      }

      throw refreshError;
    }

    return performApiRequest<T>(path, init);
  }
}
