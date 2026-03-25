export const serverEnv = {
  gatewayBaseUrl: requiredEnv("NEXT_PUBLIC_GATEWAY_BASE_URL"),
  adminServiceUrl: optionalEnv("ADMIN_SERVICE_URL"),
  adminApiKey: optionalEnv("ADMIN_API_KEY")
};

function requiredEnv(key: string): string {
  const v = process.env[key];
  if (!v || !v.trim()) {
    throw new Error(`${key} is required`);
  }
  return v.trim().replace(/\/+$/, "");
}

function optionalEnv(key: string): string | undefined {
  const v = process.env[key];
  if (!v || !v.trim()) return undefined;
  return v.trim().replace(/\/+$/, "");
}

