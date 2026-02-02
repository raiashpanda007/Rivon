import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import Cookies from 'js-cookie';

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
  profile: string
}

export interface UserState {
  userDetails: UserDetails | null;
}

export function GetUserMetaDataFromLocolStorage(): UserDetails | null {
  if (typeof window === 'undefined') return null;
  const valueFromCookie = Cookies.get("User_meta_data");
  try {
    return valueFromCookie ? JSON.parse(valueFromCookie) : null;
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
        Cookies.set("User_meta_data", userDetailsString, { expires: 7 }); // Expires in 7 days
      } catch (e) {
        console.error("UNABLE TO UPDATE USER INFO ON CLIENT SIDE :: ", e);
      }
    },
    clearUserDetails: (state) => {
      state.userDetails = null;
      try {
        Cookies.remove("User_meta_data");
      } catch (e) {
        console.error("UNABLE TO CLEAR USER INFO ON CLIENT SIDE :: ", e);
      }
    }
  },
});

export const { setUserDetails, clearUserDetails } = userSlice.actions;

export default userSlice.reducer;
