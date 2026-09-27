"use client";

import { createContext, useContext, type ReactNode } from "react";

import type { AuthUser } from "@/features/auth/types/user";

const AuthSessionContext = createContext<AuthUser | undefined>(undefined);

type AuthSessionProviderProps = {
  children: ReactNode;
  user: AuthUser;
};

export function AuthSessionProvider({
  children,
  user,
}: AuthSessionProviderProps) {
  return (
    <AuthSessionContext.Provider value={user}>
      {children}
    </AuthSessionContext.Provider>
  );
}

export function useAuthSession() {
  const user = useContext(AuthSessionContext);

  if (!user) {
    throw new Error("useAuthSession must be used inside AuthSessionProvider");
  }

  return user;
}
