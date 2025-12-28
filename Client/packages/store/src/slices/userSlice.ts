import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export enum ProviderType {
  GOOGLE = "google",
  GITHUB = "github",
  CREDENTIALS = "credentials"
}

export interface UserDetails {
  id: string;
  name: string;
  email: string;
  provider: ProviderType;
  verifiedStatus: boolean;
}

export interface UserState {
  userDetails: UserDetails | null;
}

function GetUserMetaDataFromLocolStorage(): UserDetails | null {
  if (typeof window === 'undefined') return null;
  const valueFromLocalStorage = localStorage.getItem("User_meta_data");
  try {
    return valueFromLocalStorage ? JSON.parse(valueFromLocalStorage) : null;
  } catch (e) {
    return null;
  }
}

const initialState: UserState = {
  userDetails: GetUserMetaDataFromLocolStorage(),
};

export const userSlice = createSlice({
  name: 'user',
  initialState,
  reducers: {
    setUserDetails: (state, action: PayloadAction<UserDetails>) => {
      state.userDetails = action.payload;
      try {
        const userDetailsString = JSON.stringify(action.payload);
        localStorage.setItem("User_meta_data", userDetailsString);
      } catch (e) {
        console.error("UNABLE TO UPDATE USER INFO ON CLIENT SIDE :: ", e);
      }
    },
    clearUserDetails: (state) => {
      state.userDetails = null;
      try {
        localStorage.removeItem("User_meta_data");
      } catch (e) {
        console.error("UNABLE TO CLEAR USER INFO ON CLIENT SIDE :: ", e);
      }
    }
  },
});

export const { setUserDetails, clearUserDetails } = userSlice.actions;

export default userSlice.reducer;
