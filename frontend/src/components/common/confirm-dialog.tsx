"use client";

import { useState, ReactNode } from "react";
import {
  Root,
  Trigger,
  Portal,
  Overlay,
  Content,
  Title,
  Description,
  Cancel,
  Action,
} from "@radix-ui/react-alert-dialog";
import { Button } from "@/src/components/ui/button";
import { cn } from "@/src/lib/utils";

export function ConfirmDialog({
  title,
  description,
  confirmLabel,
  cancelLabel,
  variant,
  onConfirm,
  children,
}: {
  title: string;
  description?: string;
  confirmLabel: string;
  cancelLabel?: string;
  variant?: "default" | "destructive";
  onConfirm: () => void | Promise<void>;
  children: ReactNode | ReactNode[];
}) {
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);

  return (
    <Root open={open} onOpenChange={(v) => (!busy ? setOpen(v) : null)}>
      <Trigger asChild>{children}</Trigger>
      <Portal>
        <Overlay className="fixed inset-0 z-50 bg-black/50" />
        <Content
          className={cn(
            "fixed left-1/2 top-1/2 z-50 w-[95vw] max-w-md -translate-x-1/2 -translate-y-1/2 rounded-2xl border bg-card p-6 shadow-soft",
          )}
        >
          <Title className="text-base font-semibold">
            {title}
          </Title>
          {description ? (
            <Description className="mt-2 text-sm text-muted-foreground">
              {description}
            </Description>
          ) : null}
          <div className="mt-6 flex items-center justify-end gap-2">
            <Cancel asChild>
              <Button type="button" variant="secondary" disabled={busy}>
                {cancelLabel ?? "Cancel"}
              </Button>
            </Cancel>
            <Action asChild>
              <Button
                type="button"
                variant={variant === "destructive" ? "destructive" : "default"}
                disabled={busy}
                onClick={async () => {
                  setBusy(true);
                  try {
                    await onConfirm();
                    setOpen(false);
                  } finally {
                    setBusy(false);
                  }
                }}
              >
                {confirmLabel}
              </Button>
            </Action>
          </div>
        </Content>
      </Portal>
    </Root>
  );
}
