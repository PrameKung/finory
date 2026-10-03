"use client";

import { LoaderCircleIcon } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { getGoogleSignInUrl } from "@/features/auth/api/auth-api";

type GoogleAuthButtonProps = {
  children: string;
};

function GoogleIcon() {
  return (
    <svg viewBox="0 0 24 24" className="size-5" aria-hidden="true">
      <path
        fill="#4285F4"
        d="M21.6 12.23c0-.71-.06-1.4-.18-2.07H12v3.92h5.38a4.6 4.6 0 0 1-2 3.02v2.54h3.24c1.9-1.75 2.98-4.33 2.98-7.41Z"
      />
      <path
        fill="#34A853"
        d="M12 22c2.7 0 4.98-.9 6.64-2.42l-3.24-2.54c-.9.6-2.05.96-3.4.96-2.61 0-4.83-1.77-5.62-4.14H3.04v2.62A10 10 0 0 0 12 22Z"
      />
      <path
        fill="#FBBC05"
        d="M6.38 13.86A6.02 6.02 0 0 1 6.06 12c0-.65.11-1.28.32-1.86V7.52H3.04A10 10 0 0 0 2 12c0 1.61.38 3.14 1.04 4.48l3.34-2.62Z"
      />
      <path
        fill="#EA4335"
        d="M12 6c1.47 0 2.79.5 3.83 1.5l2.88-2.88A9.65 9.65 0 0 0 12 2a10 10 0 0 0-8.96 5.52l3.34 2.62C7.17 7.77 9.39 6 12 6Z"
      />
    </svg>
  );
}

export function GoogleAuthButton({ children }: GoogleAuthButtonProps) {
  const [isRedirecting, setIsRedirecting] = useState(false);

  function startGoogleAuthentication() {
    setIsRedirecting(true);
    window.location.assign(getGoogleSignInUrl());
  }

  return (
    <Button
      type="button"
      size="lg"
      variant="outline"
      className="h-12 w-full gap-3 rounded-xl bg-background text-sm font-semibold shadow-xs hover:bg-muted/60"
      disabled={isRedirecting}
      onClick={startGoogleAuthentication}
    >
      {isRedirecting ? (
        <LoaderCircleIcon
          className="animate-spin motion-reduce:animate-none"
          aria-hidden="true"
        />
      ) : null}
      {!isRedirecting ? <GoogleIcon /> : null}
      {isRedirecting ? "Redirecting to Google…" : children}
    </Button>
  );
}
