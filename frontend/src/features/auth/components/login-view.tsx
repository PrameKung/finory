import Link from "next/link";

import { AuthPanel } from "@/features/auth/components/auth-panel";

type LoginViewProps = {
  errorMessage?: string;
};

export function LoginView({ errorMessage }: LoginViewProps) {
  return (
    <AuthPanel
      title="Welcome back"
      description="Sign in with Google to continue to your Finory dashboard."
      actionLabel="Continue with Google"
      appearance="plain"
      errorMessage={errorMessage}
      footer={
        <>
          New to Finory?{" "}
          <Link
            href="/register"
            className="font-medium text-foreground underline-offset-4 hover:underline"
          >
            Create an account
          </Link>
        </>
      }
    />
  );
}
