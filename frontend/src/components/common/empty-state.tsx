"use client";

import { Button } from "@/src/components/ui/button";
import { cn } from "@/src/lib/utils";

export function EmptyState({
  title,
  description,
  actionLabel,
  onAction,
  className
}: {
  title: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
  className?: string;
}) {
  return (
    <div className={cn("rounded-2xl border bg-card p-10 text-center shadow-soft", className)}>
      <div className="text-base font-semibold">{title}</div>
      {description ? <div className="mt-2 text-sm text-muted-foreground">{description}</div> : null}
      {actionLabel && onAction ? (
        <div className="mt-6 flex justify-center">
          <Button onClick={onAction}>{actionLabel}</Button>
        </div>
      ) : null}
    </div>
  );
}

