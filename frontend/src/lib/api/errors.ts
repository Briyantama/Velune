import type { ClientError } from "@/src/lib/api/client";

export function isClientError(e: unknown): e is ClientError {
  return (
    !!e &&
    typeof e === "object" &&
    typeof (e as any).code === "string" &&
    typeof (e as any).message === "string" &&
    typeof (e as any).status === "number"
  );
}

