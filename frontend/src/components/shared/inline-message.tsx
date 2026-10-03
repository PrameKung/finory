import { CircleAlertIcon, InfoIcon } from "lucide-react";
import type { HTMLAttributes } from "react";

import { cn } from "@/lib/utils";

type InlineMessageProps = HTMLAttributes<HTMLDivElement> & {
  variant?: "error" | "info";
};

export function InlineMessage({
  variant = "error",
  className,
  children,
  ...props
}: InlineMessageProps) {
  const Icon = variant === "error" ? CircleAlertIcon : InfoIcon;

  return (
    <div
      role={variant === "error" ? "alert" : "status"}
      className={cn(
        "flex items-start gap-2.5 rounded-lg border px-3 py-2.5 text-sm leading-5",
        variant === "error"
          ? "border-destructive/20 bg-destructive/5 text-destructive"
          : "bg-muted/40 text-muted-foreground",
        className,
      )}
      {...props}
    >
      <Icon className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
      <div className="min-w-0">{children}</div>
    </div>
  );
}
