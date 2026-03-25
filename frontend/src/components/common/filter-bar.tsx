"use client";

import { cn } from "@/src/lib/utils";

export function FilterBar({ className, children }: { className?: string; children: React.ReactNode }) {
  return (
    <div className={cn("mb-4 flex flex-col gap-3 rounded-2xl border bg-card p-4 shadow-soft md:flex-row md:items-end", className)}>
      {children}
    </div>
  );
}

