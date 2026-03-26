// BackendError represents either the new standardized envelope error shape
// (`{ error, status, requestId, ... }`) or the legacy `{ code, message, ... }` shape.
export type BackendError = {
  error?: string;
  status?: number;
  requestId?: string;

  // Legacy compatibility (some Next routes may still return this shape).
  code?: string;
  message?: string;
};

export type TokenResponse = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
};

export type MeResponse = {
  user_id: string;
  email: string;
  base_currency: string;
};

