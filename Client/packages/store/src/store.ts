import { configureStore } from "@reduxjs/toolkit";
import auth from "./reducers/userSlice.js"
export const store = configureStore({
  reducer: {
    auth: auth
  }
});

export type AppStore = typeof store;
export type RootState = ReturnType<AppStore["getState"]>;
export type AppDispatch = AppStore["dispatch"];
