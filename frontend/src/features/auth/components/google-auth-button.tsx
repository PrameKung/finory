"use client";

import { LoaderCircleIcon } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { getGoogleSignInUrl } from "@/features/auth/api/auth-api";

type GoogleAuthButtonProps = {
  children: string;
};

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
      className="w-full"
      disabled={isRedirecting}
      onClick={startGoogleAuthentication}
    >
      {isRedirecting ? (
        <LoaderCircleIcon
          className="animate-spin motion-reduce:animate-none"
          aria-hidden="true"
        />
      ) : null}
      {isRedirecting ? "Redirecting to Google…" : children}
    </Button>
  );
}
