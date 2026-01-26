"use client"
import { Button } from "@workspace/ui/components/button";
import { useRouter } from "next/navigation.js";
function SignIn() {
  const router = useRouter();

  const loginUrl = (process.env.NEXT_PUBLIC_BASE_APP_URL || "/") + "auth/login";

  return (
    <Button variant="ghost" onClick={() => window.location.href = loginUrl} className="font-medium gap-2 cursor-pointer">
      Sign in
    </Button>
  )

}



export default SignIn;
