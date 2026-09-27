import Link from "next/link";

import { AuthPanel } from "@/features/auth/components/auth-panel";

export function RegisterView() {
  return (
    <AuthPanel
      title="Create your account"
      description="Start tracking your money with a secure Google account."
      actionLabel="Sign up with Google"
      footer={
        <>
          Already have an account?{" "}
          <Link
            href="/login"
            className="font-medium text-foreground underline-offset-4 hover:underline"
          >
            Sign in
          </Link>
        </>
      }
    />
  );
}
