import type { Metadata } from "next";

import { RegisterView } from "@/features/auth/components/register-view";

export const metadata: Metadata = {
  title: "Create account",
};

export default function RegisterPage() {
  return (
    <div className="flex min-h-svh items-center justify-center px-6 py-12 sm:px-10">
      <RegisterView />
    </div>
  );
}
