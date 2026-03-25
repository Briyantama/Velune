"use client";

import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "@/src/lib/query/keys";
import { notificationsApi } from "@/src/features/notifications/services/notificationsApi";

export function useNotificationsPing() {
  return useQuery({
    queryKey: queryKeys.notificationsPing(),
    queryFn: () => notificationsApi.ping()
  });
}

