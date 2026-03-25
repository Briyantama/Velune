"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/src/lib/utils";
import { useMeQuery } from "@/src/features/auth/hooks/useMeQuery";
import { EmptyState } from "@/src/components/common/empty-state";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";

const nav = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/transactions", label: "Transactions" },
  { href: "/budgets", label: "Budgets" },
  { href: "/reports", label: "Reports" },
  { href: "/notifications", label: "Notifications" },
  { href: "/settings", label: "Settings" },
  { href: "/admin", label: "Admin" }
] as const;

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const meQ = useMeQuery();

  return (
    <div className="min-h-dvh bg-background">
      <div className="mx-auto grid max-w-7xl grid-cols-1 gap-6 px-4 py-6 md:grid-cols-[260px_1fr]">
        <aside className="h-fit rounded-2xl border bg-card p-3 shadow-soft md:sticky md:top-6">
          <div className="px-3 py-2 text-xs font-semibold tracking-wide text-muted-foreground">VELUNE</div>
          <nav className="grid gap-1 p-1">
            {nav.map((item) => {
              const active = pathname === item.href || pathname.startsWith(item.href + "/");
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "rounded-lg px-3 py-2 text-sm transition-colors hover:bg-accent",
                    active ? "bg-accent font-medium" : "text-muted-foreground"
                  )}
                >
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </aside>
        <main className="min-w-0">
          {meQ.isLoading ? <LoadingSkeleton /> : null}
          {meQ.isError ? (
            <EmptyState
              title="Couldn’t load profile"
              description="Check auth and try again."
              actionLabel="Retry"
              onAction={() => meQ.refetch()}
            />
          ) : null}
          {meQ.data ? children : null}
        </main>
      </div>
    </div>
  );
}

