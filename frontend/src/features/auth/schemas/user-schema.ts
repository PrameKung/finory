import { z } from "zod";

export const authUserSchema = z.object({
  id: z.uuid(),
  email: z.email(),
  displayName: z.string(),
  avatarUrl: z.union([z.url(), z.literal("")]),
});
