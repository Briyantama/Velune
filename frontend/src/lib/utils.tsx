import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function FieldError({ msg }: { msg?: string }) {
  if (!msg) return null;
  return <div className="text-xs text-destructive">{msg}</div>;
}

export function SafeJson(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return undefined;
  }
}

export function KV({ k, v, wide }: { k: string; v: string; wide?: boolean }) {
  return (
    <div className={wide ? "md:col-span-2" : ""}>
      <dt className="text-xs font-semibold text-muted-foreground">{k}</dt>
      <dd className="mt-1 break-words">{v}</dd>
    </div>
  );
}
