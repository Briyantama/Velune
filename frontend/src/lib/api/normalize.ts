export function camelizeKeysDeep<T>(input: unknown): T {
  return camelizeAny(input) as T;
}

function camelizeAny(v: unknown): unknown {
  if (Array.isArray(v)) return v.map(camelizeAny);
  if (!v || typeof v !== "object") return v;

  const rec = v as Record<string, unknown>;
  const out: Record<string, unknown> = {};
  for (const [k, value] of Object.entries(rec)) {
    out[toCamel(k)] = camelizeAny(value);
  }
  return out;
}

function toCamel(s: string): string {
  if (!s) return s;
  // Handles PascalCase (AccountID) and snake_case (access_token).
  if (s.includes("_")) {
    return s.toLowerCase().replace(/_([a-z0-9])/g, (_, c: string) => c.toUpperCase());
  }
  // Lowercase first char; keep common initialisms readable (ID -> id).
  const loweredFirst = s[0].toLowerCase() + s.slice(1);
  return loweredFirst.replace(/([a-z])ID\b/g, "$1Id");
}

