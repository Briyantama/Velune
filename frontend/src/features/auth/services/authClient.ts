import { camelizeKeysDeep } from "@/src/lib/api/normalize";
import { SafeJson } from "@/src/lib/utils";

export type Me = {
  userId: string;
  email: string;
  baseCurrency: string;
};

export const authClient = {
  async me(): Promise<Me> {
    const resp = await fetch("/api/auth/me", { method: "GET", headers: { Accept: "application/json" } });
    const text = await resp.text();
    const json = text ? SafeJson(text) : undefined;
    if (!resp.ok) {
      const code = (json as { code?: string })?.code ?? "HTTP_ERROR";
      const message = (json as { message?: string })?.message ?? `Request failed (${resp.status})`;
      throw { code, message, status: resp.status };
    }
    return camelizeKeysDeep<Me>(json);
  },

  async logout(): Promise<void> {
    const resp = await fetch("/api/auth/logout", { method: "POST" });
    if (!resp.ok) {
      throw { code: "LOGOUT_FAILED", message: "logout failed", status: resp.status };
    }
  }
};
