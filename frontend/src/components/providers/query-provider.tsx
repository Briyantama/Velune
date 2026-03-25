"use client";

import { MutationCache, QueryClient, QueryClientProvider, QueryCache } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { useToast } from "@/src/components/ui/toaster";
import { isClientError } from "@/src/lib/api/errors";

export function QueryProvider({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const toast = useToast();

  const clientRef = useRef<QueryClient | null>(null);
  const handled401Ref = useRef(false);
  const pathnameRef = useRef(pathname);

  useEffect(() => {
    pathnameRef.current = pathname;

    // Allow handling again after we land on a public auth page.
    if (pathname.startsWith("/login")) {
      handled401Ref.current = false;
    }
  }, [pathname]);

  const handleUnauthorized = () => {
    if (handled401Ref.current) return;
    if (pathnameRef.current.startsWith("/login")) return;

    handled401Ref.current = true;
    clientRef.current?.clear();

    // Clear httpOnly auth cookies server-side, then return user to login.
    void fetch("/api/auth/logout", { method: "POST" }).catch(() => undefined);

    toast.push({
      title: "Session expired",
      description: "Please sign in again.",
      variant: "destructive",
    });
    router.replace("/login");
  };

  const [client] = useState(() => {
    const qc = new QueryClient({
      queryCache: new QueryCache({
        onError: (error) => {
          if (isClientError(error) && error.status === 401) handleUnauthorized();
        },
      }),
      mutationCache: new MutationCache({
        onError: (error) => {
          if (isClientError(error) && error.status === 401) handleUnauthorized();
        },
      }),
      defaultOptions: {
        queries: {
          retry: (failureCount, error) => {
            const status = (error as { status?: number } | undefined)?.status;
            if (status && status >= 400 && status < 500 && status !== 429) return false;
            return failureCount < 2;
          },
          refetchOnWindowFocus: false,
        },
      },
    });

    clientRef.current = qc;
    return qc;
  });

  return (
    <QueryClientProvider client={client}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
}

