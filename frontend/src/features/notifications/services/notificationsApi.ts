import { apiGet } from "@/src/lib/api/client";

export const notificationsApi = {
  ping() {
    return apiGet<{ status: string }>("/notifications/ping");
  }
};

