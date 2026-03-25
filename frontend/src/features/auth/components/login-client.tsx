"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/src/components/ui/button";
import { Card, CardContent } from "@/src/components/ui/card";
import { Input } from "@/src/components/ui/input";
import { Label } from "@/src/components/ui/label";
import { useApiToasts } from "@/src/lib/api/toast";
import { loginSchema } from "@/src/features/auth/schema/authSchemas";
import { SafeJson, FieldError } from "@/src/lib/utils";
import { z } from "zod";

type Values = z.infer<typeof loginSchema>;

export default function LoginClient() {
  const toast = useApiToasts();
  const sp = useSearchParams();
  const next = sp.get("next") ?? "/dashboard";

  const form = useForm<Values>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });

  return (
    <Card className="w-full max-w-md">
      <CardContent className="p-6">
        <div className="text-lg font-semibold">Login</div>
        <div className="mt-1 text-sm text-muted-foreground">
          Sign in to your account.
        </div>

        <form
          className="mt-6 grid gap-4"
          onSubmit={form.handleSubmit(async (values) => {
            try {
              const resp = await fetch("/api/auth/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(values),
              });
              const text = await resp.text();
              const json = text ? SafeJson(text) : undefined;
              if (!resp.ok) {
                throw {
                  code: (json as any)?.code ?? "LOGIN_FAILED",
                  message: (json as any)?.message ?? "login failed",
                  status: resp.status,
                };
              }
              window.location.href = next;
            } catch (e) {
              toast.showError(e, "Login failed");
            }
          })}
        >
          <div className="grid gap-2">
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              autoComplete="email"
              {...form.register("email")}
            />
            <FieldError msg={form.formState.errors.email?.message} />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              {...form.register("password")}
            />
            <FieldError msg={form.formState.errors.password?.message} />
          </div>

          <Button type="submit" disabled={form.formState.isSubmitting}>
            Sign in
          </Button>

          <div className="text-sm text-muted-foreground">
            New here?{" "}
            <Link className="underline underline-offset-4" href="/register">
              Create an account
            </Link>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
