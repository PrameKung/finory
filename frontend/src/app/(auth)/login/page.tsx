import type { Metadata } from "next";

import { LoginView } from "@/features/auth/components/login-view";

export const metadata: Metadata = {
  title: "Sign in",
};

const oauthErrorMessages = {
  oauth_cancelled: "Google sign-in was cancelled. You can try again when you’re ready.",
  oauth_failed: "We couldn’t sign you in with Google. Please try again.",
} as const;

export default async function LoginPage({ searchParams }: PageProps<"/login">) {
  const error = (await searchParams).error;
  const errorMessage =
    typeof error === "string" && error in oauthErrorMessages
      ? oauthErrorMessages[error as keyof typeof oauthErrorMessages]
      : undefined;

  return <LoginView errorMessage={errorMessage} />;
}
