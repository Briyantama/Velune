import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type Me = {
  userId: string;
  email: string;
  baseCurrency: string;
};

type UserSessionState = {
  me: Me | null;
};

const initialState: UserSessionState = {
  me: null,
};

const userSessionSlice = createSlice({
  name: "userSession",
  initialState,
  reducers: {
    setMe(state, action: PayloadAction<Me>) {
      state.me = action.payload;
    },
    clearMe(state) {
      state.me = null;
    },
  },
});

export const { setMe, clearMe } = userSessionSlice.actions;
export default userSessionSlice.reducer;

