"use client";

import { LoaderCircleIcon, LogOutIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useLogout } from "@/features/auth/hooks/use-logout";
import type { AuthUser } from "@/features/auth/types/user";
import { getApiErrorMessage } from "@/lib/api/client";
import { cn } from "@/lib/utils";

type SessionIdentityProps = {
  compact?: boolean;
  user: AuthUser;
};

function getInitials(user: AuthUser) {
  const nameParts = user.displayName.trim().split(/\s+/).filter(Boolean);

  if (nameParts.length > 0) {
    return nameParts
      .slice(0, 2)
      .map((part) => part[0])
      .join("")
      .toUpperCase();
  }

  return user.email.slice(0, 1).toUpperCase();
}

export function SessionIdentity({ compact = false, user }: SessionIdentityProps) {
  const label = user.displayName.trim() || user.email;
  const logout = useLogout();

  return (
    <div className={cn("min-w-0", compact ? "" : "space-y-2")}>
      <div className={cn("flex min-w-0 items-center", compact ? "gap-0" : "gap-3")}>
        <span
          className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-semibold text-primary-foreground"
          aria-hidden="true"
        >
          {getInitials(user)}
        </span>
        {compact ? (
          <span className="sr-only">Signed in as {label}</span>
        ) : (
          <>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium text-sidebar-foreground">
                {label}
              </p>
              <p className="truncate text-xs text-sidebar-foreground/60">
                {user.email}
              </p>
            </div>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label="Sign out"
              title="Sign out"
              disabled={logout.isPending}
              onClick={() => logout.mutate()}
            >
              {logout.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : (
                <LogOutIcon aria-hidden="true" />
              )}
            </Button>
          </>
        )}
      </div>
      {!compact && logout.isError ? (
        <p role="alert" className="text-xs leading-5 text-destructive">
          {getApiErrorMessage(logout.error, {
            defaultMessage: "Could not sign out. Please try again.",
          })}
        </p>
      ) : null}
    </div>
  );
}
