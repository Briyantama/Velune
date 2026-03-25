import Link from "next/link";

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-dvh bg-background">
      <div className="mx-auto grid min-h-dvh max-w-6xl grid-cols-1 items-stretch gap-8 px-4 py-10 md:grid-cols-2">
        <div className="hidden flex-col justify-between rounded-2xl border bg-card p-8 shadow-soft md:flex">
          <div className="space-y-2">
            <div className="text-xs font-semibold tracking-wide text-muted-foreground">VELUNE</div>
            <div className="text-2xl font-semibold">Calm, accurate money tracking.</div>
            <div className="text-sm text-muted-foreground">
              Built for correctness (minor units), auditability, and safe retries.
            </div>
          </div>
          <div className="text-xs text-muted-foreground">
            <Link className="underline underline-offset-4 hover:opacity-80" href="/dashboard">
              Go to app
            </Link>
          </div>
        </div>
        <div className="flex items-center justify-center">{children}</div>
      </div>
    </div>
  );
}

