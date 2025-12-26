"use client"
import { useAppSelector } from "@workspace/store/hooks"
import SignIn from "@workspace/ui/components/Header/authOptions/SignIn";
import SignUp from "@workspace/ui/components/Header/authOptions/SignUP";
import SignOut from "@workspace/ui/components/Header/authOptions/SignOut";
import UserProfile from "@workspace/ui/components/Header/authOptions/UserProfile";

function AuthOptions() {
  const login = useAppSelector((state) => state.auth.isAuthenticated)
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
          <UserProfile />
          <SignOut />
        </>
      }
    </div>
  )
}





export default AuthOptions;
