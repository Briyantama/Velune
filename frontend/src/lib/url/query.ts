export function setIfPresent(params: URLSearchParams, key: string, value: string | undefined | null) {
  const v = (value ?? "").trim();
  if (!v) {
    params.delete(key);
    return;
  }
  params.set(key, v);
}

export function setIfBool(params: URLSearchParams, key: string, value: boolean | undefined) {
  if (value === undefined) {
    params.delete(key);
    return;
  }
  params.set(key, value ? "true" : "false");
}

export function setIfNumber(params: URLSearchParams, key: string, value: number | undefined | null) {
  if (value === undefined || value === null || Number.isNaN(value)) {
    params.delete(key);
    return;
  }
  params.set(key, String(value));
}
