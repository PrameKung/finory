import Link from "next/link";

import { AuthPanel } from "@/features/auth/components/auth-panel";

export function LoginView() {
  return (
    <AuthPanel
      title="Welcome back"
      description="Sign in to continue to your Finory dashboard."
      actionLabel="Continue with Google"
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
