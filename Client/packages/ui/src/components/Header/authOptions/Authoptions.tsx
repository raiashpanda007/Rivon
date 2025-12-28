"use client"
import { useState, useEffect } from "react";
import SignIn from "@workspace/ui/components/Header/authOptions/SignIn";
import SignUp from "@workspace/ui/components/Header/authOptions/SignUP";
import SignOut from "@workspace/ui/components/Header/authOptions/SignOut";
import UserProfile from "@workspace/ui/components/Header/authOptions/UserProfile";
import WalletIcon from "@workspace/ui/components/Header/authOptions/Wallet";
import { useSelector } from "react-redux";
import type { RootState } from "@workspace/store";

function AuthOptions() {
  const userDetails = useSelector((state: RootState) => state.user.userDetails);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const login = mounted && userDetails !== null;
  return (
    <div className="flex items-center gap-3">
      {!login &&
        <>
          <SignIn />
          <SignUp />
        </>
      }
      {login &&
        <>
          <UserProfile url={userDetails.profile} />
          <SignOut />
          <WalletIcon />
        </>
      }
    </div>
  )
}





export default AuthOptions;
