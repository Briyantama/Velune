"use client";

import * as React from "react";

import { cn } from "@/src/lib/utils";

type AlertVariant = "success" | "warning" | "error" | "info";

export function AlertCard({
  variant,
  title,
  description,
  className,
  actions,
}: {
  variant: AlertVariant;
  title: string;
  description?: string;
  className?: string;
  actions?: React.ReactNode;
}) {
  const stylesByVariant: Record<AlertVariant, string> = {
    success: "border-success/30 bg-success/10 text-success-foreground",
    warning: "border-warning/30 bg-warning/10 text-warning-foreground",
    error: "border-error/30 bg-error/10 text-error-foreground",
    info: "border-info/30 bg-info/10 text-info-foreground",
  };

  return (
    <div
      role="alert"
      className={cn(
        "rounded-2xl border p-4 shadow-soft",
        stylesByVariant[variant],
        className
      )}
    >
      <div className="flex flex-col gap-2">
        <div className="text-sm font-semibold">{title}</div>
        {description ? <div className="text-sm opacity-90">{description}</div> : null}
        {actions ? <div className="pt-1">{actions}</div> : null}
      </div>
    </div>
  );
}

