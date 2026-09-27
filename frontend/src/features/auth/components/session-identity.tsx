import type { AuthUser } from "@/features/auth/types/user";
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

  return (
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
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-sidebar-foreground">
            {label}
          </p>
          <p className="truncate text-xs text-sidebar-foreground/60">
            {user.email}
          </p>
        </div>
      )}
    </div>
  );
}
