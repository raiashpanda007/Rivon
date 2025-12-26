import { useState } from "react";
import SignIn from "@workspace/ui/components/Header/authOptions/SignIn";
import SignUp from "@workspace/ui/components/Header/authOptions/SignUP";
import SignOut from "@workspace/ui/components/Header/authOptions/SignOut";
import UserProfile from "@workspace/ui/components/Header/authOptions/UserProfile";

function AuthOptions() {
  const [login, setLogin] = useState<boolean>(true);
  return (
    <div className="w-1/6 h-full flex items-center justify-evenly">
      {!login &&
        <>
          <SignUp />
          <SignIn />
        </>
      }
      {login &&
        <>
          <SignOut />
          <UserProfile />
        </>
      }
    </div>
  )
}





export default AuthOptions;
