"use client";

import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "@/src/lib/query/keys";
import { authClient, type Me } from "@/src/features/auth/services/authClient";

export function useMeQuery() {
  return useQuery<Me>({
    queryKey: queryKeys.me(),
    queryFn: () => authClient.me(),
  });
}

