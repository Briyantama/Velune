"use client";

import { Button } from "@/src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/src/components/ui/dialog";
import type { StorageMode } from "@/src/store/slices/consentSlice";

export function ConsentModal({
  open,
  onAccept,
  onReject,
}: {
  open: boolean;
  onAccept: () => void;
  onReject: () => void;
}) {
  return (
    <Dialog open={open} onOpenChange={() => undefined}>
      <DialogContent>
        <DialogTitle>Storage preference</DialogTitle>
        <DialogDescription>
          We store your sign-in session so you don&apos;t have to log in every time.
          Choose how you want it stored on this device.
        </DialogDescription>

        <div className="mt-6 grid gap-3">
          <div className="rounded-2xl border bg-card p-4">
            <div className="text-sm font-semibold">Use cookies</div>
            <div className="mt-1 text-sm text-muted-foreground">
              Uses secure cookies (recommended). Helps keep you signed in.
            </div>
            <div className="mt-4">
              <Button type="button" onClick={onAccept}>
                Allow cookies
              </Button>
            </div>
          </div>

          <div className="rounded-2xl border bg-card p-4">
            <div className="text-sm font-semibold">Use local storage</div>
            <div className="mt-1 text-sm text-muted-foreground">
              Does not store auth tokens in cookies. Uses local storage in your browser.
            </div>
            <div className="mt-4">
              <Button type="button" variant="secondary" onClick={onReject}>
                Use local storage instead
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

