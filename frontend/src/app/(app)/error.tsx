"use client";

import { ErrorState } from "@/components/shared/error-state";

type AppErrorProps = {
  error: Error & { digest?: string };
  retry: () => void;
};

export default function AppError({ retry }: AppErrorProps) {
  return (
    <ErrorState
      description="We could not load this page. Try again, or return later if the problem continues."
      onRetry={retry}
    />
  );
}
