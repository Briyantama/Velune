"use client";

import * as React from "react";
import * as ToastPrimitive from "@radix-ui/react-toast";
import { cn } from "@/src/lib/utils";

type ToastState = {
  id: string;
  title?: string;
  description?: string;
  variant?: "default" | "destructive";
};

const ToastCtx = React.createContext<{
  push: (t: Omit<ToastState, "id">) => void;
} | null>(null);

export function useToast() {
  const ctx = React.useContext(ToastCtx);
  if (!ctx) throw new Error("useToast must be used within <Toaster />");
  return ctx;
}

export function Toaster({ children }: { children?: React.ReactNode }) {
  const [toasts, setToasts] = React.useState<ToastState[]>([]);

  const push = React.useCallback((t: Omit<ToastState, "id">) => {
    const id = crypto.randomUUID();
    setToasts((prev) => [...prev, { id, ...t }]);
  }, []);

  return (
    <ToastCtx.Provider value={{ push }}>
      <ToastPrimitive.Provider swipeDirection="right" duration={3500}>
        {children}
        {toasts.map((t) => (
          <ToastPrimitive.Root
            key={t.id}
            className={cn(
              "group pointer-events-auto relative flex w-full max-w-md items-start gap-3 overflow-hidden rounded-lg border p-4 shadow-soft",
              t.variant === "destructive" ? "border-destructive/40 bg-destructive/10" : "bg-card"
            )}
            onOpenChange={(open) => {
              if (!open) setToasts((prev) => prev.filter((x) => x.id !== t.id));
            }}
          >
            <div className="grid gap-1">
              {t.title ? <ToastPrimitive.Title className="text-sm font-semibold">{t.title}</ToastPrimitive.Title> : null}
              {t.description ? (
                <ToastPrimitive.Description className="text-sm text-muted-foreground">{t.description}</ToastPrimitive.Description>
              ) : null}
            </div>
          </ToastPrimitive.Root>
        ))}
        <ToastPrimitive.Viewport className="fixed bottom-0 right-0 z-50 flex w-full max-w-md flex-col gap-2 p-4 outline-none" />
      </ToastPrimitive.Provider>
    </ToastCtx.Provider>
  );
}

