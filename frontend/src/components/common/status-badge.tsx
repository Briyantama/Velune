"use client";

import { cn } from "@/src/lib/utils";

type StatusVariant = "neutral" | "success" | "warning" | "error" | "info";

export function StatusBadge({
  variant,
  label,
  className,
}: {
  variant: StatusVariant;
  label: string;
  className?: string;
}) {
  const stylesByVariant: Record<StatusVariant, string> = {
    neutral: "bg-muted/40 text-muted-foreground border-border/60",
    success: "bg-success/15 text-success-foreground border-success/30",
    warning: "bg-warning/15 text-warning-foreground border-warning/30",
    error: "bg-error/15 text-error-foreground border-error/30",
    info: "bg-info/15 text-info-foreground border-info/30",
  };

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold",
        stylesByVariant[variant],
        className,
      )}
    >
      {label}
    </span>
  );
}
