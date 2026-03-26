import { configureStore } from "@reduxjs/toolkit";
import authReducer from "@/src/store/slices/authSlice";
import consentReducer from "@/src/store/slices/consentSlice";
import userSessionReducer from "@/src/store/slices/userSessionSlice";

export const store = configureStore({
  reducer: {
    auth: authReducer,
    consent: consentReducer,
    userSession: userSessionReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

