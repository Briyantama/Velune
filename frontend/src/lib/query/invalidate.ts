import type { QueryClient } from "@tanstack/react-query";

export function invalidateMany(qc: QueryClient, prefixes: string[]) {
  for (const p of prefixes) {
    qc.invalidateQueries({ queryKey: [p] });
  }
}
