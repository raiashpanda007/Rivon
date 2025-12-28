"use client"
import { useEffect } from "react";
import { useDispatch } from "react-redux";
import ApiCaller, { RequestType } from "@workspace/api-caller";
import { setUserDetails, clearUserDetails, ProviderType } from "@workspace/store";
import { useAppSelector } from "@workspace/store";

export default function AuthCheck() {
  const dispatch = useDispatch();
  const userDetails = useAppSelector((state) => state.user.userDetails);

  useEffect(() => {
    const checkAuth = async () => {
      // If user details are already in the store (loaded from localStorage), skip the API call
      if (userDetails) {
        return;
      }

      const response = await ApiCaller<undefined, { id: string, name: string, email: string, provider: string, verified: boolean, profile: string | undefined }>({
        requestType: RequestType.GET,
        paths: ["api", "rivon", "auth", "me"],
        retry: false,

      });

      if (response.ok) {
        dispatch(setUserDetails({
          id: response.response.data.id,
          name: response.response.data.name,
          email: response.response.data.email,
          provider: response.response.data.provider as ProviderType,
          verifiedStatus: response.response.data.verified,
          profile: response.response.data.profile ?? ""
        }));
      } else {
        dispatch(clearUserDetails());
      }
    };

    checkAuth();
  }, [dispatch, userDetails]);

  return null;
}
