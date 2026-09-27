import { authUserSchema } from "@/features/auth/schemas/user-schema";
import type { AuthUser } from "@/features/auth/types/user";
import { ApiError, apiRequest, getApiUrl } from "@/lib/api/client";

const authPath = "/api/v1/auth";

let refreshPromise: Promise<void> | undefined;

export const authQueryKeys = {
  all: ["auth"] as const,
  currentUser: ["auth", "current-user"] as const,
};

export function getGoogleSignInUrl() {
  return getApiUrl(`${authPath}/google`);
}

export async function refreshSession() {
  refreshPromise ??= apiRequest<void>(`${authPath}/refresh`, {
    method: "POST",
  }).finally(() => {
    refreshPromise = undefined;
  });

  return refreshPromise;
}

async function requestCurrentUser(): Promise<AuthUser> {
  const data = await apiRequest<unknown>(`${authPath}/me`, {
    cache: "no-store",
  });

  return authUserSchema.parse(data);
}

export async function getCurrentUser(): Promise<AuthUser | null> {
  try {
    return await requestCurrentUser();
  } catch (error) {
    if (!(error instanceof ApiError) || error.status !== 401) {
      throw error;
    }
  }

  try {
    await refreshSession();
    return await requestCurrentUser();
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return null;
    }

    throw error;
  }
}

export function logout() {
  return apiRequest<void>(`${authPath}/logout`, {
    method: "POST",
  });
}
