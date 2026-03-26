import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type StorageMode = "cookie" | "localStorage";

type ConsentState = {
  status: "unknown" | "consented";
  storageMode: StorageMode | null;
};

const initialState: ConsentState = {
  status: "unknown",
  storageMode: null,
};

const consentSlice = createSlice({
  name: "consent",
  initialState,
  reducers: {
    setConsent(state, action: PayloadAction<{ storageMode: StorageMode }>) {
      state.status = "consented";
      state.storageMode = action.payload.storageMode;
    },
    clearConsent(state) {
      state.status = "unknown";
      state.storageMode = null;
    },
  },
});

export const { setConsent, clearConsent } = consentSlice.actions;
export default consentSlice.reducer;

