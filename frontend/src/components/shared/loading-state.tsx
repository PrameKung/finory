import { LoaderCircleIcon } from "lucide-react";

import { cn } from "@/lib/utils";

type LoadingStateProps = {
  title?: string;
  description?: string;
  className?: string;
};

export function LoadingState({
  title = "Loading",
  description,
  className,
}: LoadingStateProps) {
  return (
    <div
      role="status"
      aria-busy="true"
      aria-live="polite"
      className={cn(
        "flex min-h-64 w-full flex-col items-center justify-center gap-3 rounded-xl border border-dashed bg-card px-6 py-12 text-center",
        className,
      )}
    >
      <span className="flex size-11 items-center justify-center rounded-full bg-muted text-muted-foreground">
        <LoaderCircleIcon
          className="size-5 animate-spin motion-reduce:animate-none"
          aria-hidden="true"
        />
      </span>
      <div className="space-y-1">
        <h2 className="font-heading text-base font-medium">{title}</h2>
        {description ? (
          <p className="max-w-sm text-sm text-muted-foreground">
            {description}
          </p>
        ) : null}
      </div>
    </div>
  );
}
