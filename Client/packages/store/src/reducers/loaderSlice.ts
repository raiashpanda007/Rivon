import { createSlice, PayloadAction } from "@reduxjs/toolkit";

interface LoaderState {
  state: boolean;
  heading?: string;
  message?: string;
  app?: 'bet' | 'exchange' | 'both';
}

const initialState: LoaderState = {
  state: false,
};

interface SetLoaderPayload {
  state: boolean;
  heading?: string;
  message?: string;
  app?: 'bet' | 'exchange' | 'both';
}
const loaderSlice = createSlice({
  name: "loader",
  initialState,
  reducers: {
    setLoader(state, action: PayloadAction<SetLoaderPayload>) {
      state.state = action.payload.state;
      state.heading = action.payload.heading;
      state.message = action.payload.message;
      state.app = action.payload.app;
    },
    clearLoader(state) {
      state.state = false;
      state.heading = undefined;
      state.message = undefined;
      state.app = undefined;
    }
  },
});

export const { setLoader, clearLoader } = loaderSlice.actions;
export default loaderSlice.reducer;
