"use client"
import { Button } from "@workspace/ui/components/button";
import { FaSignOutAlt } from "react-icons/fa";
import ApiCaller, { RequestType } from "@workspace/api-caller"
import { useState } from "react";
import { useSelector, useDispatch } from "react-redux";
import { RootState, clearUserDetails } from "@workspace/store";
import { useRouter } from "next/navigation";

import { ShowResponseToast } from "../../toast";

function SignOut() {
  const [loading, isLoading] = useState(false);
  const userDetails = useSelector((state: RootState) => state.user.userDetails);
  const dispatch = useDispatch();
  const router = useRouter();
  const BASE_APP_URL = process.env.NEXT_PUBLIC_BASE_APP_URL

  async function SignOutUser(id: string) {
    if (!BASE_APP_URL) {
      throw Error("PLEASE PROVIDE BASE APP URL");
    }
    if (!id) {
      ShowResponseToast({
        heading: "Error",
        message: "User ID not found.",
        statusCode: 400,
        type: "ERROR"
      });
      return;
    }

    isLoading(true);
    try {
      await ApiCaller<{ id: string }, undefined>({
        paths: ["api", "rivon", "auth", "credentials", "signout"],
        requestType: RequestType.DELETE,
        body: { id },
        retry: false
      });

      ShowResponseToast({
        heading: "Success",
        message: "Signed out successfully.",
        statusCode: 200,
        type: "SUCESS"
      });

      dispatch(clearUserDetails());
      router.replace(BASE_APP_URL)
    } catch (error: any) {
      ShowResponseToast({
        heading: "Error",
        message: error.message || "Failed to sign out.",
        statusCode: 500,
        type: "ERROR"
      });
    } finally {
      isLoading(false);
    }
  };

  return (
    <Button
      variant={"destructive"}
      className="font-body font-semibold cursor-pointer"
      onClick={() => SignOutUser(userDetails?.id || "")}
      disabled={loading}
    >
      {loading ? "Signing Out..." : "Sign Out"} <FaSignOutAlt className="ml-2" />
    </Button>
  )

}

export default SignOut;
