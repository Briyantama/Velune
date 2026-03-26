import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type AuthStatus =
  | "unknown"
  | "unauthenticated"
  | "hydrating"
  | "authenticated";

type AuthState = {
  status: AuthStatus;
  lastError?: { code: string; message: string };
};

const initialState: AuthState = {
  status: "unknown",
};

const authSlice = createSlice({
  name: "auth",
  initialState,
  reducers: {
    setStatus(state, action: PayloadAction<AuthStatus>) {
      state.status = action.payload;
      state.lastError = undefined;
    },
    setError(state, action: PayloadAction<{ code: string; message: string }>) {
      state.lastError = action.payload;
      state.status = "unauthenticated";
    },
    clearAuthState(state) {
      state.status = "unauthenticated";
      state.lastError = undefined;
    },
  },
});

export const { setStatus, setError, clearAuthState } = authSlice.actions;
export default authSlice.reducer;

