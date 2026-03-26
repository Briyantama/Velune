import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().trim().email(),
  password: z.string().min(1)
});

export const registerSchema = z.object({
  email: z.string().trim().email(),
  password: z.string().min(8).max(72),
  confirmPassword: z.string().min(8).max(72),
  baseCurrency: z.string().trim().length(3).default("IDR")
}).refine((data) => data.password === data.confirmPassword, {
  path: ["confirmPassword"],
  message: "Passwords do not match"
});
