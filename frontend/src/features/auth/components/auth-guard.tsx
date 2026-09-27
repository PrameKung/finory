"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { ErrorState } from "@/components/shared/error-state";
import { LoadingState } from "@/components/shared/loading-state";
import { useCurrentUser } from "@/features/auth/hooks/use-current-user";
import { AuthSessionProvider } from "@/features/auth/providers/auth-session-provider";
import { getApiErrorMessage } from "@/lib/api/client";

type AuthGuardProps = {
  children: ReactNode;
};

export function AuthGuard({ children }: AuthGuardProps) {
  const router = useRouter();
  const currentUser = useCurrentUser();

  useEffect(() => {
    if (currentUser.isSuccess && currentUser.data === null) {
      router.replace("/login");
    }
  }, [currentUser.data, currentUser.isSuccess, router]);

  if (currentUser.isError) {
    return (
      <div className="flex min-h-svh items-center justify-center bg-muted/30 p-6">
        <ErrorState
          className="max-w-lg bg-background"
          title="Unable to verify your session"
          description={getApiErrorMessage(currentUser.error, {
            defaultMessage:
              "We could not confirm that you are signed in. Please try again.",
          })}
          onRetry={() => void currentUser.refetch()}
        />
      </div>
    );
  }

  if (currentUser.isPending || currentUser.data === null) {
    return (
      <div className="flex min-h-svh items-center justify-center bg-muted/30 p-6">
        <LoadingState
          className="max-w-lg bg-background"
          title={currentUser.data === null ? "Returning to sign in" : "Checking your session"}
          description="Your protected workspace will open once your session is verified."
        />
      </div>
    );
  }

  return (
    <AuthSessionProvider user={currentUser.data}>
      {children}
    </AuthSessionProvider>
  );
}
