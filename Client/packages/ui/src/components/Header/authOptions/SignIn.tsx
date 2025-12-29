"use client"
import { Button } from "@workspace/ui/components/button";
import { useRouter } from "next/navigation.js";
function SignIn() {
  const router = useRouter();
  return (
    <Button variant="ghost" onClick={() => router.push("/auth/login")} className="font-medium gap-2 cursor-pointer">
      Sign in
    </Button>
  )

}



export default SignIn;
