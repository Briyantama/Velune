"use client";

import { createContext, useCallback, useContext, useState } from "react";
import type { ReactNode } from "react";
import {
  Provider,
  Root,
  Title,
  Description,
  Viewport,
} from "@radix-ui/react-toast";
import { cn } from "@/src/lib/utils";

type ToastState = {
  id: string;
  title?: string;
  description?: string;
  variant?: "default" | "destructive";
};

const ToastCtx = createContext<{
  push: (t: Omit<ToastState, "id">) => void;
} | null>(null);

export function useToast() {
  const ctx = useContext(ToastCtx);
  if (!ctx) throw new Error("useToast must be used within <Toaster />");
  return ctx;
}

export function Toaster({ children }: { children?: ReactNode }) {
  const [toasts, setToasts] = useState<ToastState[]>([]);

  const push = useCallback((t: Omit<ToastState, "id">) => {
    const id = crypto.randomUUID();
    setToasts((prev) => [...prev, { id, ...t }]);
  }, []);

  return (
    <ToastCtx.Provider value={{ push }}>
      <Provider swipeDirection="right" duration={3500}>
        {children}
        {toasts.map((t) => (
          <Root
            key={t.id}
            className={cn(
              "group pointer-events-auto relative flex w-full max-w-md items-start gap-3 overflow-hidden rounded-lg border p-4 shadow-soft",
              t.variant === "destructive"
                ? "border-destructive/40 bg-destructive/10"
                : "bg-card",
            )}
            onOpenChange={(open) => {
              if (!open) setToasts((prev) => prev.filter((x) => x.id !== t.id));
            }}
          >
            <div className="grid gap-1">
              {t.title ? (
                <Title className="text-sm font-semibold">{t.title}</Title>
              ) : null}
              {t.description ? (
                <Description className="text-sm text-muted-foreground">
                  {t.description}
                </Description>
              ) : null}
            </div>
          </Root>
        ))}
        <Viewport className="fixed bottom-0 right-0 z-50 flex w-full max-w-md flex-col gap-2 p-4 outline-none" />
      </Provider>
    </ToastCtx.Provider>
  );
}
