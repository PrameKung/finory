import type { z } from "zod";

import type { authUserSchema } from "@/features/auth/schemas/user-schema";

export type AuthUser = z.infer<typeof authUserSchema>;
