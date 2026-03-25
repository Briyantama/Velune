"use client";

import Link from "next/link";
import { Resolver, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/src/components/ui/button";
import { Card, CardContent } from "@/src/components/ui/card";
import { Input } from "@/src/components/ui/input";
import { Label } from "@/src/components/ui/label";
import { useApiToasts } from "@/src/lib/api/toast";
import { registerSchema } from "@/src/features/auth/schema/authSchemas";
import { FieldError, SafeJson } from "@/src/lib/utils";

type Values = z.infer<typeof registerSchema>;

export default function RegisterClient() {
  const toast = useApiToasts();
  const form = useForm<Values>({
    resolver: zodResolver(registerSchema) as Resolver<Values, any, Values>,
    defaultValues: { email: "", password: "", baseCurrency: "USD" },
  });

  return (
    <Card className="w-full max-w-md">
      <CardContent className="p-6">
        <div className="text-lg font-semibold">Create account</div>
        <div className="mt-1 text-sm text-muted-foreground">
          Start tracking with a clean ledger.
        </div>

        <form
          className="mt-6 grid gap-4"
          onSubmit={form.handleSubmit(async (values) => {
            try {
              const resp = await fetch("/api/auth/register", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(values),
              });
              const text = await resp.text();
              const json = text ? SafeJson(text) : undefined;
              if (!resp.ok) {
                throw {
                  code: (json as any)?.code ?? "REGISTER_FAILED",
                  message: (json as any)?.message ?? "register failed",
                  status: resp.status,
                };
              }
              window.location.href = "/dashboard";
            } catch (e) {
              toast.showError(e, "Register failed");
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
              autoComplete="new-password"
              {...form.register("password")}
            />
            <FieldError msg={form.formState.errors.password?.message} />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="baseCurrency">Base currency</Label>
            <Input
              id="baseCurrency"
              placeholder="USD"
              {...form.register("baseCurrency")}
            />
            <FieldError msg={form.formState.errors.baseCurrency?.message} />
          </div>

          <Button type="submit" disabled={form.formState.isSubmitting}>
            Create account
          </Button>

          <div className="text-sm text-muted-foreground">
            Already have an account?{" "}
            <Link className="underline underline-offset-4" href="/login">
              Sign in
            </Link>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
