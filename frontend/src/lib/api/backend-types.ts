export type BackendError = {
  code: string;
  message: string;
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

